package main

import (
	"net/http"

	"github.com/joshua-zingale/ucr-learning-services/tree/master/infrastructure/taxis/internal/web"
)

func main() {

	mux := web.NewMux()

	http.ListenAndServe("0.0.0.0:3456", mux)
}
