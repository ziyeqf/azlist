package main

import (
	"encoding/json"
	"fmt"
	"os"

	azlog "github.com/Azure/azure-sdk-for-go/sdk/azcore/log"
	"github.com/hashicorp/go-hclog"
	"github.com/magodo/azlist/azlist"

	"github.com/urfave/cli/v2"
)

func main() {
	var (
		flagTenantId           string
		flagClientId           string
		flagClientSecret       string
		flagClientCertPath     string
		flagClientCertPassword string
		flagEnvironment        string
		flagSubscriptionId     string
		flagRecursive          bool
		flagWithBody           bool
		flagIncludeManaged     bool
		flagParallelism        int
		flagPrintError         bool
		flagVerbose            bool
	)

	app := &cli.App{
		Name:      "azlist",
		Version:   getVersion(),
		Usage:     "List Azure resources by an Azure Resource Graph `where` predicate",
		UsageText: "azlist [option] <ARG where predicate>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "client-id",
				EnvVars:     []string{"AZLIST_CLIENT_ID", "ARM_CLIENT_ID"},
				Usage:       "The client id",
				Destination: &flagClientId,
			},
			&cli.StringFlag{
				Name:        "client-secret",
				EnvVars:     []string{"AZLIST_CLIENT_SECRET", "ARM_CLIENT_SECRET"},
				Usage:       "The client secret",
				Destination: &flagClientSecret,
			},
			&cli.StringFlag{
				Name:        "client-certificate-path",
				EnvVars:     []string{"AZLIST_CLIENT_CERTIFICATE_PATH", "ARM_CLIENT_CERTIFICATE_PATH"},
				Usage:       "The client certificate path",
				Destination: &flagClientCertPath,
			},
			&cli.StringFlag{
				Name:        "client-certificate-password",
				EnvVars:     []string{"AZLIST_CLIENT_CERTIFICATE_PASSWORD", "ARM_CLIENT_CERTIFICATE_PASSWORD"},
				Usage:       "The client certificate password",
				Destination: &flagClientCertPassword,
			},
			&cli.StringFlag{
				Name:        "tenant-id",
				EnvVars:     []string{"AZLIST_TENANT_ID", "ARM_TENANT_ID"},
				Usage:       "The tenant id",
				Destination: &flagTenantId,
			},
			&cli.StringFlag{
				Name:        "env",
				EnvVars:     []string{"AZLIST_ENV"},
				Usage:       `The environment. Can be one of "public", "china", "usgovernment".`,
				Destination: &flagEnvironment,
				Value:       "public",
			},
			&cli.StringFlag{
				Name:        "subscription-id",
				EnvVars:     []string{"AZLIST_SUBSCRIPTION_ID", "ARM_SUBSCRIPTION_ID"},
				Aliases:     []string{"s"},
				Required:    true,
				Usage:       "The subscription id",
				Destination: &flagSubscriptionId,
			},
			&cli.BoolFlag{
				Name:        "recursive",
				Aliases:     []string{"r"},
				EnvVars:     []string{"AZLIST_RECURSIVE"},
				Usage:       "Recursively list child resources of the query result",
				Destination: &flagRecursive,
			},
			&cli.BoolFlag{
				Name:        "with-body",
				EnvVars:     []string{"AZLIST_WITH_BODY"},
				Aliases:     []string{"b"},
				Usage:       "Print each resource's body",
				Destination: &flagWithBody,
			},
			&cli.BoolFlag{
				Name:        "include-managed",
				Aliases:     []string{"m"},
				EnvVars:     []string{"AZLIST_INCLUDE_MANAGED"},
				Usage:       "Include resource whose lifecycle is managed by others",
				Destination: &flagIncludeManaged,
			},
			&cli.IntFlag{
				Name:        "parallelism",
				EnvVars:     []string{"AZLIST_PARALLELISM"},
				Aliases:     []string{"p"},
				Usage:       "Limit the number of parallel operations to list resources",
				Value:       10,
				Destination: &flagParallelism,
			},
			&cli.BoolFlag{
				Name:        "print-error",
				Aliases:     []string{"e"},
				EnvVars:     []string{"AZLIST_PRINT_ERROR"},
				Usage:       "Print errors received during listing resources",
				Destination: &flagPrintError,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				EnvVars:     []string{"AZLIST_VERBOSE"},
				Usage:       "Print verbose output",
				Destination: &flagVerbose,
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() == 0 {
				return fmt.Errorf("No ARG where predicate specified")
			}
			if ctx.NArg() > 1 {
				return fmt.Errorf("More than one where predicates specified")
			}

			if flagVerbose {
				logger := hclog.New(&hclog.LoggerOptions{
					Name:  "azlist",
					Level: hclog.Debug,
				}).StandardLogger(&hclog.StandardLoggerOptions{
					InferLevels: true,
				})
				azlist.SetLogger(logger)

				os.Setenv("AZURE_SDK_GO_LOGGING", "all")
				azlog.SetListener(func(cls azlog.Event, msg string) {
					logger.Printf("[DEBUG] %s: %s\n", cls, msg)
				})
			}

			opt := azlist.Option{
				SubscriptionId:     flagSubscriptionId,
				TenantID:           flagTenantId,
				ClientID:           flagClientId,
				ClientSecret:       flagClientSecret,
				ClientCertPath:     flagClientCertPath,
				ClientCertPassword: flagClientCertPassword,
				Env:                flagEnvironment,
				Parallelism:        flagParallelism,
				Recursive:          flagRecursive,
				IncludeManaged:     flagIncludeManaged,
			}

			result, err := azlist.List(ctx.Context, ctx.Args().First(), opt)
			if err != nil {
				return err
			}

			if flagPrintError {
				if len(result.Errors) != 0 {
					fmt.Println("Listing errors:")
					for _, err := range result.Errors {
						fmt.Printf("\t%v\n", err)
					}
					fmt.Println()
				}
			}

			for _, res := range result.Resources {
				fmt.Println(res.Id)
				if flagWithBody {
					b, _ := json.MarshalIndent(res.Properties, "", "  ")
					fmt.Println(string(b))
				}
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
