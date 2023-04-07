package api

type SASLConfig struct {
	Type       string
	Username   string `yaml:"username,omitempty" json:"username,omitempty" doc:"SASL username"`
	SecretPath string `yaml:"secretPath,omitempty" json:"secretPath,omitempty" doc:"path to the user secret (password)"`
}

type SASLTypeEnum struct {
	Plain       string `yaml:"plain" json:"plain" doc:"Plain SASL"`
	ScramSHA512 string `yaml:"scramSHA512" json:"scramSHA512" doc:"SCRAM/SHA512 SASL"`
}

func SASLTypeName(operation string) string {
	return GetEnumName(SASLTypeEnum{}, operation)
}
