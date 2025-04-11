package models

import "github.com/wangfenjin/mojito/pkg/migrations"

func init() {
	migrations.RegisterModel("user", "1.0.0", &UserV1{}, nil)
	migrations.RegisterModel("item", "1.0.1", &ItemV1{}, nil)
	migrations.RegisterModel("user", "1.0.2", &UserV2{}, &UserV1{})
}
