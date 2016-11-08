package koalasdk

//url参数
type UrlParams map[string]string
//body参数
type BodyParams map[string]interface{}

/*
api 信息
 */
type KoalaApi struct {
	//api basic info
	Api string
	ApiVersion string

	//need login user token info
	UserId uint64
	UserRole uint8
	Token string

	//request time stamp
	Timestamp uint64
	//user agent
	UserAgent string

	//url 其他参数
	UrlMap UrlParams

	BodyMap BodyParams

}

/*
风控参数列表,放在head部分
 */
type KoalaDS struct {

	Imei string;

	Imsi string;

	Sim string;

	RouteId string;

	RouteMac string;

	BsId string;

	RegionCode string;

	MobileMac string;

	UtaId string;

	UtdId string;

	MbRooted bool;

	AppInSim bool;

}

type KoalaResponse struct{
	Code int
	Data string
	Message string
}

