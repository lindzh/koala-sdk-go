package koalasdk

import (
	"time"
	"errors"
	"encoding/json"
	"html"
	"net/http"
	"bytes"
	"net/url"
	"io/ioutil"
	"strings"
	"os"
	"mime/multipart"
	"fmt"
	"github.com/axgle/mahonia"
	"compress/gzip"
	"strconv"
)

type KoalaClient struct {
	//基本信息
	AppKey string
	AppSec string
	GatewayUrl string

	//签名
	signer Signer
	//风控参数
	ds *KoalaDS

	httpClient *http.Client
}

func (client *KoalaClient)Init(){
	var sign H5Signer = H5Signer{}
	client.signer = sign

	//抓包调试使用
	proxy := func(r *http.Request) (*url.URL, error) {
		return url.Parse("http://127.0.0.1:8888")
	}

	transport := &http.Transport{Proxy: proxy}

	client.httpClient = &http.Client{Transport:transport}

	fmt.Println("hahah--------started")
}

func toString(data interface{}) string {
	switch data.(type) {
	case []byte :
		bb := data.([]byte)
		ll := len(bb)
		ls := string(ll)
		return strings.Join([]string{"b[",ls,"]"},"")
	case os.File:
		ff := data.(os.File)
		stat,err := ff.Stat()
		if err != nil {
			panic(err)
		}
		ll := stat.Size()
		ls := string(ll)
		return strings.Join([]string{"b[",ls,"]"},"")
	default:
		return data.(string)
	}
}

func (client *KoalaClient) buildUrl(urlParams *UrlParams) string {
	buf := bytes.NewBufferString("")
	buf.WriteString(client.GatewayUrl)
	buf.WriteString("?")
	for k,v := range *urlParams {
		buf.WriteString(url.QueryEscape(k))
		buf.WriteString("=")
		buf.WriteString(url.QueryEscape(v))
		buf.WriteString("&")
	}
	return string(buf.Bytes())
}

/**
生成url参数
 */
func (client *KoalaClient) buildUrlMap(api *KoalaApi) map[string]string {

	data := make(map[string]string)
	//先将用户定义的url放入参数中
	if api.UrlMap != nil {
		for k,v:= range api.UrlMap{
			data[k] = v
		}
	}
	//覆盖系统设置参数,不能修改
	data[URL_API_NAME] = api.Api
	data[URL_API_VERSION] = api.ApiVersion
	data[URL_APP_KEY] = client.AppKey
	data[URL_APP_VERSION] = APP_VERSION
	data[URL_TIMESTAMP] = strconv.FormatUint(api.Timestamp,10)

	if api.UserId>0 {
		data[URL_TOKEN] = api.Token
		data[URL_USER_ID] = string(api.UserId)
		data[URL_USER_ROLE] = string(api.UserRole)
	}
	var osType int64 = int64(OS_TYPE)
	data[URL_OS_TYPE] = strconv.FormatInt(osType,10)
	data[URL_TTID] = TTID
	data[URL_MOBILE_TYPE] = MOBILE_TYPE
	data[URL_HW_ID] = "go-hwid-1.7.3"
	data[URL_OS_VERSION] = OS_VERSION

	return data
}

/*
生成head map
 */
func (client *KoalaClient) buildHeadMap(api *KoalaApi) map[string]string {
	head := make(map[string]string)
	head[HEAD_USER_AGENT_KEY] = USER_AGENT

	if client.ds != nil {
		data,_ := json.Marshal(client.ds)
		dstr := string(data)
		//url encode
		head[HEAD_KOPDS_KEY] = html.EscapeString(dstr)
	}
	return head
}

/*
生成签名签字符串
 */
func (client *KoalaClient) buildSignMap(api *KoalaApi,urlmap *UrlParams) map[string]string {
	data := make(map[string]string)
	for k,v := range *urlmap {
		data[k] = v
	}

	if api.BodyMap != nil {
		for k,v := range api.BodyMap {
			if v != nil {
				data[k] = toString(v)
			}
		}
	}
	return data
}



/**
请求发送
 */
func (client *KoalaClient)Request(api *KoalaApi) (*KoalaResponse,error){
	//client param validate
	if client.AppKey== "" || client.AppSec== "" {
		err := errors.New("client must set appkey and appsec for request")
		return nil,err
	}
	if api.Api== "" || api.ApiVersion== ""{
		err := errors.New("api and api version can't be null")
		return nil,err

	}

	//
	if api.Timestamp < 10000 {
		api.Timestamp = uint64(time.Now().Unix()*1000)
	}
	fmt.Println(api.Timestamp)

	//生成签名
	urlMap := client.buildUrlMap(api)
	umap := UrlParams(urlMap)
	signMap := client.buildSignMap(api,&umap)
	fmt.Println(signMap)
	signParams := SignParams(signMap)
	signBefore := client.signer.SignBefore(&signParams,client.AppSec)
	sign := client.signer.Sign(signBefore)

	fmt.Println("sign:"+sign+",signBefore:"+signBefore)
	//生成url参数列表
	urlMap[URL_SIGN] = sign

	umap = UrlParams(urlMap)

	requestUrl := client.buildUrl(&umap)

	head := client.buildHeadMap(api)

	return client.doRequest(requestUrl,head,api.BodyMap)
}

/**
判断是multipart还是普通jsonbody
 */
func (client *KoalaClient) hasFile(body *BodyParams) bool {
	if body == nil {
		return false
	}
	bb := map[string]interface{}(*body)
	for _,v := range bb {
		if v != nil {
			switch v.(type) {
			case []byte :
				return true
			case os.File :
				return true
			default:
				continue
			}
		}
	}
	return false
}

/**
multipart body
 */
func buildFileBody(body *BodyParams) []byte{
	if(body== nil){
		ebody := url.QueryEscape("{}")
		return []byte(ebody)
	}
	buf := bytes.NewBuffer(nil)
	wr := multipart.NewWriter(buf)
	bb := map[string]interface{}(*body)
	for k,v := range bb {
		if v != nil {
			switch v.(type) {
			case []byte :
				bw,err := wr.CreateFormField(k)
				if err != nil {
					panic(err)
				}
				bs := v.([]byte)
				bw.Write(bs)
			case os.File :
				file := v.(os.File)
				fw,err := wr.CreateFormFile(k,file.Name())
				if err != nil {
					panic(err)
				}
				bdata,berr := ioutil.ReadFile(file.Name())
				if berr != nil {
					panic(berr)
				}
				_,err2 := fw.Write(bdata)
				if err2 != nil {
					panic(err2)
				}
			default:
				wr.WriteField(k,v.(string))
			}
		}
	}
	return buf.Bytes()
}

/**
普通body
 */
func buildSimpleBody(body *BodyParams) []byte{
	if(body== nil){
		ebody := url.QueryEscape("{}")
		return []byte(ebody)
	}
	data,err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	jdata := string(data)

	ebody := url.QueryEscape(jdata)
	return []byte(ebody)
}


func (client *KoalaClient) doRequest(requestUrl string,head UrlParams,body BodyParams) (*KoalaResponse,error){
	var bodyData []byte
	hasFile:= client.hasFile(&body)
	if hasFile{
		bodyData = buildFileBody(&body)
	}else{
		bodyData = buildSimpleBody(&body)
	}

	//buf := ioutil.NopCloser(bytes.NewReader(bodyData))
	fmt.Println(string(bodyData))
	buf := strings.NewReader(string(bodyData))
	//buf := bytes.NewReader(bodyData)
	//POST必须大写
	req,err := http.NewRequest("POST",requestUrl,buf)
	if err != nil {
		return nil,err
	}
	if head != nil {
		for k,v := range head {
			req.Header.Add(k,v)
		}
	}

	fmt.Println(hasFile)

	if !hasFile {
		req.Header.Add("Content-Type","application/json")
	}else {
		req.Header.Add("Content-Type","multipart/form-data")
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36")

	req.Header.Add("Accept", "*/*")

	req.Header.Add("Accept-Encoding", "gzip, deflate")

	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,en;q=0.6")

	req.Header.Add("Cache-Control", "no-cache")

	resp,err := client.httpClient.Do(req)


	if err != nil {
		return nil,err
	}

	if resp.StatusCode != 200 {
		errstr := strings.Join([]string{"http response code wrong ", string(resp.StatusCode)},"")
		dd,_:= ioutil.ReadAll(resp.Body)
		cccss := string(dd)
		fmt.Println("request failed "+cccss)
		err2 := errors.New(errstr)
		return nil,err2
	}
	return client.parseResponse(resp)

}

/*
 body gzip 压缩格式,需要解压缩
 */
func (client *KoalaClient) parseResponse(resp *http.Response) (*KoalaResponse,error) {
	fmt.Println("hahaha-----resp")

	defer resp.Body.Close()

	reader,_:= gzip.NewReader(resp.Body)

	data,err := ioutil.ReadAll(reader)
	if err != nil {
		return nil,err
	}



	datastr := string(data)

	fmt.Println(datastr)

	enc := mahonia.NewDecoder("UTF-8")

	ss := enc.ConvertString(datastr)

	fmt.Println(ss)

	ret := make(map[string]interface{})
	jsonerr := json.Unmarshal(data,ret)
	if jsonerr != nil {
		errmsg := strings.Join([]string{"parese json to map error",datastr},",")
		err2 := errors.New(errmsg)
		return nil,err2
	}

	code := ret["code"]
	if code == nil {
		errmsg := strings.Join([]string{"invalid response",datastr},",")
		err2 := errors.New(errmsg)
		return nil,err2
	}
	cd := code.(int)
	rst := KoalaResponse{}
	rst.Code = cd
	msg := ret["message"]
	if msg != nil {
		rst.Message = msg.(string)
	}

	dd := ret["data"]
	if dd != nil {
		ds,derr := json.Marshal(dd)
		if derr != nil {
			errmsg := strings.Join([]string{"parse json data error",datastr},",")
			err2 := errors.New(errmsg)
			return nil,err2

		}else{
			rst.Data = string(ds)
		}

	}

	return &rst,nil
}

