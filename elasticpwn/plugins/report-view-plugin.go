package EPPlugins

import (
	"fmt"
	"log"
	"net/http"
)

type ReportViewPlugin struct {
	ReportDirectory string
	ServeAtPort     string
}

func (rvp *ReportViewPlugin) Run() {
	log.Println(fmt.Sprintf("Report being served at http://localhost:%v", rvp.ServeAtPort))
	err := http.ListenAndServe(fmt.Sprintf(":%v", rvp.ServeAtPort), http.FileServer(http.Dir(rvp.ReportDirectory)))
	if err != nil {
		log.Fatal(err)
	}
}
