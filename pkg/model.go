package mongo

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TLSEntity struct {
	CaCert        string `json:"ca_cert" bson:"ca_cert" yaml:"ca_cert" mapstructure:"ca_cert"`
	ClientCert    string `json:"client_cert" bson:"client_cert" yaml:"client_cert" mapstructure:"client_cert"`
	ClientCertKey string `json:"client_cert_key" bson:"client_cert_key" yaml:"client_cert_key" mapstructure:"client_cert_key"`
}

type ConfigEntity struct {
	Tls TLSEntity `json:"tls" bson:"tls" yaml:"tls" mapstructure:"tls"`

	Account  string `json:"account" bson:"account" yaml:"account" mapstructure:"account"`
	Password string `json:"password" bson:"password" yaml:"password" mapstructure:"password"`

	Address  string `json:"address" yaml:"address" mapstructure:"address"`
	Database string `json:"database" yaml:"database" mapstructure:"database"`

	Mode bool `json:"mode" yaml:"mode" mapstructure:"mode"` // Mode is true cluster

	MaxOpenConnects int `json:"max_open_connects" bson:"max_open_connects" yaml:"max_open_connects" mapstructure:"max_open_connects"`
	MaxIdleConnects int `json:"max_idle_connects" bson:"max_idle_connects" yaml:"max_idle_connects" mapstructure:"max_idle_connects"`
	ConnMaxLifeTime int `json:"conn_max_life_time" bson:"conn_max_life_time" yaml:"conn_max_life_time" mapstructure:"conn_max_life_time"`

	LoggerEnable bool `json:"logger_enable" bson:"logger_enable" yaml:"logger_enable" mapstructure:"logger_enable"`
}

type TableEntity struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	CreatedAt int64              `json:"created_at" bson:"created_at"`
	UpdatedAt int64              `json:"updated_at" bson:"updated_at"`
	DeletedAt *int64             `json:"deleted_at" bson:"deleted_at,omitempty"`
}
