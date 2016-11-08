package koalasdk

import (
	"sort"
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"crypto/sha1"
	"fmt"
)

type H5Signer struct {

}

func (h5Signer H5Signer) SignBefore(param *SignParams,sec string) string {
	//排序keys
	keys := make([]string,0)

	for k,_ := range *param {
		keys = append(keys,k)
	}
	sort.Strings(keys)
	skeys := sort.StringSlice(keys)
	sort.Sort(sort.Reverse(skeys))

	//组装字符
	buf := bytes.NewBufferString("")
	buf.WriteString(sec)

	for _,k := range skeys {
		buf.WriteString(k)
		value := (*param)[k]
		buf.WriteString(value)
	}

	buf.WriteString(sec)

	return buf.String()
}


func (h5Signer H5Signer) Sign(str string) string {
	fmt.Println(str)
	//base64 encode
	strbin := []byte(str)
	encode := base64.StdEncoding.EncodeToString(strbin)


	//sha1
	hash := sha1.New()

	data := []byte(encode)
	hash.Write(data)
	sha1Bin := hash.Sum(nil)

	ddss := string(sha1Bin)
	fmt.Println(ddss)


	//hex
	return hex.EncodeToString(sha1Bin)
}
