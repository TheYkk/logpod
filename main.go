package main

import (
	"context"
	"flag"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	kubeconfig = flag.String("kubeconfig", os.Getenv("KUBECONFIG"), "absolute path to the kubeconfig file")
	annotation = Getenv("WATCH_ANNOTATION", "timestamp")
	namespaces = Getenv("WATCH_NAMESPACES", "")
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Msg("Start operator")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	ctx := context.Background()
	watcher, err := clientset.CoreV1().Pods(v1.NamespaceAll).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Make a resultchan
	ch := watcher.ResultChan()

	// Loop events of pod
	for event := range ch {
		pod, ok := event.Object.(*v1.Pod)
		if !ok {
			log.Fatal().Msg("unexpected type")
		}
		if namespaces != "" {
			if !strings.Contains(pod.Namespace, namespaces) {
				continue
			}
		}

		if event.Type == watch.Added {
			if time.Now().Sub(pod.CreationTimestamp.Time).Seconds() > 10 {
				continue
			}
			log.Info().Str("name", pod.Name).Msg("Pod created")

			time.Sleep(time.Second * 5)

			pod, err = clientset.CoreV1().Pods(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
			if err != nil {
				log.Error().Err(err)
			}

			pod.SetAnnotations(map[string]string{annotation: strconv.FormatInt(time.Now().Unix(), 10)})

			_, err = clientset.CoreV1().Pods(pod.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
			if err != nil {
				log.Error().Err(err)
			}
			log.Info().Msg("Pod annotation updated")
		}
	}
}
func Getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
