package controller

// Response represents the response code list after updating the microgateway
// maps [project -> response code]
//
// a_com
//   Updated
// b_com
//   Failed
// c_com
//   Deleted
//
type Response map[string]ResponseType

type ResponseType int

const (
	Failed  = ResponseType(0)
	Updated = ResponseType(1)
	Deleted = ResponseType(2)
)
