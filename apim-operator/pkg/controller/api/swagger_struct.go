package api

//SwaggerStruct contains the structure of basic yaml
//Openapi string `yaml:"openapi"`
type SwaggerStruct struct {
	
	Servers []struct {
		URL string `json:"url" yaml:"url"`
	} `json:"servers" yaml:"servers"`
	Info struct {
		Description    string `json:"description" yaml:"description"`
		Version        string `json:"version" yaml:"version"`
		Title          string `json:"title" yaml:"title"`
		TermsOfService string `json:"termsOfService" yaml:"termsOfService"`
		Contact        struct {
			Email string `json:"email" yaml:"email"`
		} `json:"contact" yaml:"contact"`
		License struct {
			Name string `json:"name" yaml:"name"`
			URL  string `json:"url" yaml:"url"`
		} `json:"license" yaml:"license"`
	} `json:"info" yaml:"info"`
	Tags []struct {
		Name         string `json:"name" yaml:"name"`
		Description  string `json:"description" yaml:"description"`
		ExternalDocs struct {
			Description string `json:"description" yaml:"description"`
			URL         string `json:"url" yaml:"url"`
		} `json:"externalDocs,omitempty" aml:"externalDocs,omitempty"`
	} `json:"tags" yaml:"tags"`
	XMgwBasePath            string `json:"x-mgw-basePath" yaml:"x-mgw-basePath"`
	XMgwProductionEndpoints struct {
		Urls []string `json:"urls" yaml:"urls"`
	} `json:"x-mgw-production-endpoints" yaml:"x-mgw-production-endpoints"`
	Paths struct {
		PetFindByStatus struct {
			Get struct {
				Tags        []string `json:"tags" yaml:"tags"`
				Summary     string   `json:"summary" yaml:"summary"`
				Description string   `json:"description" yaml:"description"`
				OperationID string   `json:"operationId" yaml:"operationId"`
				Parameters  []struct {
					Name        string `json:"name" yaml:"name"`
					In          string `json:"in" yaml:"in"`
					Description string `json:"description" yaml:"description"`
					Required    bool   `json:"required" yaml:"required"`
					Explode     bool   `json:"explode" yaml:"explode"`
					Schema      struct {
						Type  string `json:"type" yaml:"type"`
						Items struct {
							Type    string   `json:"type" yaml:"type"`
							Enum    []string `json:"enum" yaml:"enum"`
							Default string   `json:"default" yaml:"default"`
						} `json:"items" yaml:"items"`
					} `json:"schema" yaml:"schema"`
				} `json:"parameters" yaml:"parameters"`
				Responses struct {
					Num200 struct {
						Description string `json:"description" yaml:"description"`
						Content     struct {
							ApplicationXML struct {
								Schema struct {
									Type  string `json:"type" yaml:"type"`
									Items struct {
										Ref string `json:"$ref" yaml:"$ref"`
									} `json:"items" yaml:"items"`
								} `json:"schema" yaml:"schema"`
							} `json:"application/xml" yaml:"application/xml"`
							ApplicationJSON struct {
								Schema struct {
									Type  string `json:"type" yaml:"type"`
									Items struct {
										Ref string `json:"$ref" yaml:"$ref"`
									} `json:"items" yaml:"items"`
								} `json:"schema" yaml:"schema"`
							} `json:"application/json" yaml:"application/json"`
						} `json:"content" yaml:"content"`
					} `json:"200" yaml:"200"`
					Num400 struct {
						Description string `json:"description" yaml:"description"`
					} `json:"400" yaml:"400"`
				} `json:"responses" yaml:"responses"`
			} `json:"get" yaml:"get"`
		} `json:"/pet/findByStatus" yaml:"/pet/findByStatus"`
		PetPetID struct {
			Get struct {
				Tags        []string `json:"tags" yaml:"tags"`
				Summary     string   `json:"summary" yaml:"summary"`
				Description string   `json:"description" yaml:"description"`
				OperationID string   `json:"operationId" yaml:"operationId"`
				Parameters  []struct {
					Name        string `json:"name" yaml:"name"`
					In          string `json:"in" yaml:"in"`
					Description string `json:"description" yaml:"description"`
					Required    bool   `json:"required" yaml:"required"`
					Schema      struct {
						Type   string `json:"type" yaml:"type"`
						Format string `json:"format" yaml:"format"`
					} `json:"schema" yaml:"schema"`
				} `json:"parameters" yaml:"parameters"`
				Responses struct {
					Num200 struct {
						Description string `json:"description" yaml:"description"`
						Content     struct {
							ApplicationXML struct {
								Schema struct {
									Ref string `json:"$ref" yaml:"$ref"`
								} `json:"schema" yaml:"schema"`
							} `json:"application/xml" yaml:"application/xml"`
							ApplicationJSON struct {
								Schema struct {
									Ref string `json:"$ref" yaml:"$ref"`
								} `json:"schema" yaml:"schema"`
							} `json:"application/json" yaml:"application/json"`
						} `json:"content" yaml:"content"`
					} `json:"200" yaml:"200"`
					Num400 struct {
						Description string `json:"description" yaml:"description"`
					} `json:"400" yaml:"400"`
					Num404 struct {
						Description string `json:"description" yaml:"description"`
					} `json:"404" yaml:"404"`
				} `json:"responses" yaml:"responses"`
			} `json:"get" yaml:"get"`
		} `json:"/pet/{petId}" yaml:"/pet/{petId}"`
	} `json:"paths" yaml:"paths"`
	Components struct {
		Schemas struct {
			Category struct {
				Type       string `json:"type" yaml:"type"`
				Properties struct {
					ID struct {
						Type   string `json:"type" yaml:"type"`
						Format string `json:"format" yaml:"format"`
					} `json:"id" yaml:"id"`
					Name struct {
						Type string `json:"type" yaml:"type"`
					} `json:"name" yaml:"name"`
				} `json:"properties" yaml:"properties"`
				XML struct {
					Name string `json:"name" yaml:"name"`
				} `json:"xml" yaml:"xml"`
			} `json:"Category" yaml:"Category"`
			Tag struct {
				Type       string `json:"type" yaml:"type"`
				Properties struct {
					ID struct {
						Type   string `json:"type" yaml:"type"`
						Format string `json:"format" yaml:"format"`
					} `json:"id" yaml:"id"`
					Name struct {
						Type string `json:"type" yaml:"type"`
					} `json:"name" yaml:"name"`
				} `json:"properties" yaml:"properties"`
				XML struct {
					Name string `json:"name" yaml:"name"`
				} `json:"xml" yaml:"xml"`
			} `json:"Tag" yaml:"Tag"`
			Pet struct {
				Type       string   `json:"type" yaml:"type"`
				Required   []string `json:"required" yaml:"required"`
				Properties struct {
					ID struct {
						Type   string `json:"type" yaml:"type"`
						Format string `json:"format" yaml:"format"`
					} `json:"id" yaml:"id"`
					Category struct {
						Ref string `json:"$ref" yaml:"$ref"`
					} `json:"category" yaml:"category"`
					Name struct {
						Type    string `json:"type" yaml:"type"`
						Example string `json:"example" yaml:"example"`
					} `json:"name" yaml:"name"`
					PhotoUrls struct {
						Type string `json:"type" yaml:"type"`
						XML  struct {
							Name    string `json:"name" yaml:"name"`
							Wrapped bool   `json:"wrapped" yaml:"wrapped"`
						} `json:"xml" yaml:"xml"`
						Items struct {
							Type string `json:"type" yaml:"type"`
						} `json:"items" yaml:"items"`
					} `json:"photoUrls" yaml:"photoUrls"`
					Tags struct {
						Type string `json:"type" yaml:"type"`
						XML  struct {
							Name    string `json:"name" yaml:"name"`
							Wrapped bool   `json:"wrapped" yaml:"wrapped"`
						} `json:"xml" yaml:"xml"`
						Items struct {
							Ref string `json:"$ref" yaml:"$ref"`
						} `json:"items" yaml:"items"`
					} `json:"tags" yaml:"tags"`
					Status struct {
						Type        string   `json:"type" yaml:"type"`
						Description string   `json:"description" yaml:"description"`
						Enum        []string `json:"enum" yaml:"enum"`
					} `json:"status" yaml:"status"`
				} `json:"properties" yaml:"properties"`
				XML struct {
					Name string `json:"name" yaml:"name"`
				} `json:"xml" yaml:"xml"`
			} `json:"Pet" yaml:"Pet"`
		} `json:"schemas" yaml:"schemas"`
		RequestBodies struct {
			Pet struct {
				Content struct {
					ApplicationJSON struct {
						Schema struct {
							Ref string `json:"$ref" yaml:"$ref"`
						} `json:"schema" yaml:"schema"`
					} `json:"application/json yaml:"application/json"`
					ApplicationXML struct {
						Schema struct {
							Ref string `json:"$ref" yaml:"$ref"`
						} `json:"schema" yaml:"schema"`
					} `json:"application/xml" yaml:"application/xml"`
				} `json:"content" yaml:"content"`
				Description string `json:"description" yaml:"description"`
				Required    bool   `json:"required" yaml:"required"`
			} `json:"Pet" yaml:"Pet"`
		} `json:"requestBodies" yaml:"requestBodies"`
	} `json:"components" yaml:"components"`
}
