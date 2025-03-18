package aws_lib

type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SubjectToken_GcpWorkloadIdentityFederation struct {
	Url     string         `json:"url"`
	Method  string         `json:"method"`
	Headers []KeyValuePair `json:"headers"`
}
