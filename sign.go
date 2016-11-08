package koalasdk

//basic signer define

type SignParams map[string]string

type Signer interface {

SignBefore(param *SignParams,sec string) string

Sign(str string) string

}
