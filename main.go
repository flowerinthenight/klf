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
	// main parent (root) command
	rootCmd = &cobra.Command{
		Use:   "klf",
		Short: "kubectl logs follower for multiple pods",
		Long:  `A simple wrapper for [kubectl logs -f] for multiple pods.`,
	}
)

func main() {
	log.SetFlags(0)

	rootCmd.AddCommand(
		TailCmd(),
	)

	err := rootCmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}

func TailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tail <svc>",
		Short: "tail a k8s service logs",
		Long:  `Tail a k8s service for logs.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				log.Println("No service name provided")
				os.Exit(1)
			}

			type _spec struct {
				Selector json.RawMessage
			}

			type _svc struct {
				Spec _spec `json:"spec"`
			}

			c := exec.Command("kubectl", "get", "svc", args[0], "-o", "json")
			out, _ := c.CombinedOutput()

			var svc _svc
			var sel map[string]string
			var lbl string

			json.Unmarshal(out, &svc)
			json.Unmarshal(svc.Spec.Selector, &sel)
			for k, v := range sel {
				lbl += k + "=" + v + ","
			}

			c = exec.Command("kubectl", "get", "pod", "-l", strings.TrimRight(lbl, ","), "-o", "json")
			out, _ = c.CombinedOutput()
			log.Print(string(out))
		},
	}

	return cmd
}
