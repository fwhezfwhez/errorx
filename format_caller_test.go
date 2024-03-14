package errorx

import (
	"fmt"
	"testing"
)

func TestFormatCaller(t *testing.T) {
	control()
}

func control() {
	if e := service(); e != nil {
		fmt.Printf("handle user info error, err=%s \n", Wrap(service()).Error())
		//for _, v := range GetStack(Wrap(service())) {
		//	fmt.Printf("%s \n", v)
		//}
	}
}

func service() error {
	e := serviceB()
	return Wrap(e)
}

func serviceB() error {
	return WrapContext(model(), map[string]interface{}{
		"app_id":     "akkkk",
		"incr_times": 15,
	})
}
func model() error {
	//return NewFromStringf("query user db err=%s", db().Error())

	return WrapContext(db(), map[string]interface{}{
		"game_id": 7,
		"name":    "fengtao",
	})
}

func db() error {
	return fmt.Errorf("nil return")
}
