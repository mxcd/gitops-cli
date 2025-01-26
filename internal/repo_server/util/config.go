package util

import "github.com/mxcd/go-config/config"

func InitConfig() error {
	err := config.LoadConfig([]config.Value{
		config.String("LOG_LEVEL").NotEmpty().Default("info"),
		config.Int("PORT").Default(8080),

		config.Bool("DEV").Default(false),

		config.String("API_BASE_URL").Default("/api/v1"),

		config.String("GITOPS_REPOSITORY").NotEmpty(),
		config.String("GITOPS_REPOSITORY_BRANCH").NotEmpty().Default("main"),
		config.Bool("GITOPS_REPOSITORY_NO_STRICT_HOST_KEY_CHECKING").Default(false),
		config.String("GITOPS_REPOSITORY_HOST_KEY").Default(""),

		config.String("GITOPS_REPOSITORY_BASICAUTH").Sensitive(),
		config.String("GITOPS_REPOSITORY_SSH_KEY").Sensitive(),
		config.String("GITOPS_REPOSITORY_SSH_KEY_FILE").Sensitive(),
		config.String("GITOPS_REPOSITORY_SSH_KEY_PASSPHRASE").Sensitive(),

		config.StringArray("API_KEYS").NotEmpty().Sensitive(),
	})
	return err
}
