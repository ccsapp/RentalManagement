package environment

import "os"

func setCarServerUrl(carServerUrl string) {
	_ = os.Setenv(envCarServerUrl, carServerUrl)
}
