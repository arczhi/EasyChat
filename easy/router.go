package easy

import (
	"net/http"
)

/*
a defined struct
*
*/
type router struct {
	http.ServeMux
	// halt
}
