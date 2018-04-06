package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "klf",
		Short: "kubectl logs follower for multiple pods",
		Long:  `A simple wrapper for [kubectl logs] for multiple pods.`,
	}

	addprefix bool
)

func main() {
	log.SetFlags(0)
	rootCmd.AddCommand(tailcmd())
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}

func errexit(err error) {
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
}

func tailcmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tail <svc|dep> [extra-args]",
		Short: "tail a k8s service/deployment logs",
		Long:  `Tail a k8s service/deployment for logs.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				log.Println("No service/deployment name provided")
				os.Exit(1)
			}

			var (
				err error
				c   *exec.Cmd
				out []byte
				sel map[string]string
				lbl string
			)

			switch args[0] {
			case "svc":
				type _spec struct {
					Selector json.RawMessage
				}

				type _svc struct {
					Spec _spec `json:"spec"`
				}

				var svc _svc

				log.Printf("Retrieving service '%v' information...", args[1])
				c = exec.Command("kubectl", "get", "svc", args[1], "-o", "json")
				out, err = c.CombinedOutput()
				errexit(err)

				err = json.Unmarshal(out, &svc)
				errexit(err)

				err = json.Unmarshal(svc.Spec.Selector, &sel)
				errexit(err)
			case "dep":
				type _match struct {
					Match json.RawMessage `json:"matchLabels"`
				}

				type _spec struct {
					Selector _match `json:"selector"`
				}

				type _dep struct {
					Spec _spec `json:"spec"`
				}

				var dep _dep

				log.Printf("Retrieving deployment '%v' information...", args[1])
				c = exec.Command("kubectl", "get", "deployment", args[1], "-o", "json")
				out, err = c.CombinedOutput()
				errexit(err)

				err = json.Unmarshal(out, &dep)
				errexit(err)

				err = json.Unmarshal(dep.Spec.Selector.Match, &sel)
				errexit(err)
			default:
				log.Println("Invalid input")
				os.Exit(1)
			}

			for k, v := range sel {
				lbl += k + "=" + v + ","
			}

			lbl = strings.TrimRight(lbl, ",")
			log.Printf("Getting all pods with label(s) '%v'...", lbl)

			c = exec.Command("kubectl", "get", "pod", "-l", strings.TrimRight(lbl, ","), "-o", "json")
			out, err = c.CombinedOutput()
			errexit(err)

			type _metadata struct {
				Name string `json:"name"`
			}

			type _items struct {
				Metadata _metadata `json:"metadata"`
			}

			type _pod struct {
				Items []_items `json:"items"`
			}

			var pod _pod

			err = json.Unmarshal(out, &pod)
			errexit(err)

			log.Println("Pods to tail:", pod.Items)
			log.Println("Log command for each: kubectl logs -f {pod}", strings.Join(args[2:], " "))
			for _, p := range pod.Items {
				logargs := make([]string, 0)
				logargs = append(logargs, "logs")
				logargs = append(logargs, "-f")
				logargs = append(logargs, p.Metadata.Name)
				for _, x := range args[2:] {
					logargs = append(logargs, x)
				}

				prefix := p.Metadata.Name
				lc := exec.Command("kubectl", logargs...)
				outpipe, _ := lc.StdoutPipe()
				errpipe, _ := lc.StderrPipe()
				err = lc.Start()
				errexit(err)

				go func() {
					outscan := bufio.NewScanner(outpipe)
					for {
						chk := outscan.Scan()
						if !chk {
							break
						}

						stxt := outscan.Text()
						if addprefix {
							log.Println("["+prefix+"]", stxt)
						} else {
							log.Println(stxt)
						}
					}
				}()

				go func() {
					errscan := bufio.NewScanner(errpipe)
					for {
						chk := errscan.Scan()
						if !chk {
							break
						}

						stxt := errscan.Text()
						if addprefix {
							log.Println("["+prefix+"]", stxt)
						} else {
							log.Println(stxt)
						}
					}
				}()
			}

			errs := make(chan error)

			go func() {
				c := make(chan os.Signal)
				signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
				sig := fmt.Errorf("%s", <-c)
				errs <- sig
			}()

			<-errs
		},
	}

	cmd.Flags().BoolVar(&addprefix, "add-prefix", addprefix, "add pod name as log prefix")
	return cmd
}
