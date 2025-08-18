package cmd

import (
	"fmt"

	"github.com/geektr-cloud/fcddns/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type JwtClaims struct {
	Domain string `json:"domain"`
	Host   string `json:"host"`
	IP     string `json:"ip"`
	jwt.RegisteredClaims
}

var jwtCmd = &cobra.Command{
	Use: "jwt <subcommand>",
}

var jwtSignCmd = &cobra.Command{
	Use:   "sign <domain> <host>",
	Short: "sign message with jwt secret",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		secret := viper.GetString("jwt-secret")
		domain := args[0]
		host := args[1]

		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, JwtClaims{
			Domain: domain,
			Host:   host,
		}).SignedString([]byte(secret))

		if err != nil {
			fmt.Printf("failed to sign message: %v\n", err)
			return
		}

		fmt.Println(token)
	},
}

var jwtVerifyCmd = &cobra.Command{
	Use:   "verify <token>",
	Short: "verify message with jwt secret",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		secret := viper.GetString("jwt-secret")
		tokenStr := utils.ExpandStdin(args[0])

		claims := JwtClaims{}
		token, _ := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		fmt.Printf("host: %s\n", claims.Host)
		fmt.Printf("domain: %s\n", claims.Domain)
		fmt.Printf("ip: %s\n", claims.IP)
		fmt.Printf("valid: %v\n", token.Valid)
	},
}

func init() {
	jwtSignCmd.Flags().StringP("jwt-secret", "s", "", "jwt secret")
	viper.BindPFlags(jwtSignCmd.Flags())
	jwtCmd.AddCommand(jwtSignCmd)

	jwtVerifyCmd.Flags().StringP("jwt-secret", "s", "", "jwt secret")
	viper.BindPFlags(jwtVerifyCmd.Flags())
	jwtCmd.AddCommand(jwtVerifyCmd)

	rootCmd.AddCommand(jwtCmd)
}
