package controller

// Response represents the response code list after updating the microgateway
// maps [project -> response code]
//
// example1_com
//   200
// example2_com
//   500
//
type Response map[string]int
