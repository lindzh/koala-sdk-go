package koalasdk

import "fmt"
import "testing"

func TestClient(t *testing.T){

	client := KoalaClient{AppKey:"e2c2ab71e88548978956f426e9529d",AppSec:"77bb4062f36f405194665645508ff73",GatewayUrl:"http://ttest.gagaga.com"}
	client.Init()
	api := &KoalaApi{}
	api.Api = "kop.time"
	api.ApiVersion = "1.0"
	api.BodyMap = make(map[string]interface{})
	api.BodyMap["echo"] = "56666"

	resp,err := client.Request(api)

	if err != nil {
		fmt.Println("hahahah--------error")
		fmt.Println("hahahah--------error")
		fmt.Print(err)
	}else{
		fmt.Println(resp.Code)
		fmt.Println(resp.Data)
		fmt.Println("message:"+resp.Message)
	}
}
