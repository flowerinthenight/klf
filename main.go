package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "klf",
		Short: "kubectl logs follower for multiple pods",
		Long:  `A simple wrapper for [kubectl logs] for multiple pods.`,
	}
)

func main() {
	log.SetFlags(0)
	rootCmd.AddCommand(TailCmd())
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}

func TailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tail <svc|dep> [extra-args]",
		Short: "tail a k8s service/deployment logs",
		Long:  `Tail a k8s service/deployment for logs.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				log.Println("No service/deployment name provided")
				os.Exit(1)
			}

			type _spec struct {
				Selector json.RawMessage
			}

			type _svc struct {
				Spec _spec `json:"spec"`
			}

			var target string

			switch args[0] {
			case "svc":
				c := exec.Command("kubectl", "get", "svc", args[1], "-o", "json")
				out, _ := c.CombinedOutput()
			case "dep":
				c := exec.Command("kubectl", "get", "deployment", args[1], "-o", "json")
				out, _ := c.CombinedOutput()
			default:
				log.Println("Invalid input")
				os.Exit(1)
			}

			var svc _svc
			var sel map[string]string
			var lbl string

			json.Unmarshal(out, &svc)
			json.Unmarshal(svc.Spec.Selector, &sel)
			for k, v := range sel {
				lbl += k + "=" + v + ","
			}

			log.Println("lbl", lbl)
			os.Exit(0)

			c = exec.Command("kubectl", "get", "pod", "-l", strings.TrimRight(lbl, ","), "-o", "json")
			out, _ = c.CombinedOutput()
			log.Print(string(out))
		},
	}

	return cmd
}
