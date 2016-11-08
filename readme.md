### go koala sdk

> how to use


example
```go
	client := KoalaClient{AppKey:"e2c2ab71e88548978956f426e9529d",AppSec:"77bb4062f36f405194665645508ff73",GatewayUrl:"http://ttest.gagaga.com"}
	client.Init()
	api := &KoalaApi{}
	api.Api = "kop.time"
	api.ApiVersion = "1.0"
	api.BodyMap = make(map[string]interface{})
	api.BodyMap["echo"] = "56666"

	resp,err := client.Request(api)

```