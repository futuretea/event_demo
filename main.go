package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	clientset "k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"log"
	"time"
)

type EventClient struct {
	nodeName       string
	client         typedcorev1.CoreV1Interface
	eventNamespace string
}

func NewEventClient(cfg *rest.Config, eventNamespace, nodeName string) *EventClient {
	t := EventClient{}
	t.client = clientset.NewForConfigOrDie(cfg).CoreV1()
	t.eventNamespace = eventNamespace
	t.nodeName = nodeName
	return &t
}

func eventHandler(e *v1.Event) {
	fmt.Println(
		e.Name,
		e.Namespace,
		e.Type,
		e.Reason,
		e.InvolvedObject.Name,
		e.Message)
}

func getEventRecorder(c typedcorev1.CoreV1Interface, namespace, nodeName, source string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.V(9).Infof)
	eventBroadcaster.StartEventWatcher(eventHandler)
	recorder := eventBroadcaster.NewRecorder(legacyscheme.Scheme, v1.EventSource{Component: source, Host: nodeName})
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: c.Events(namespace)})
	return recorder
}

func getNodeRef(namespace, nodeName string) *v1.ObjectReference {
	return &v1.ObjectReference{
		Kind:      "Node",
		Name:      nodeName,
		UID:       types.UID(nodeName),
		Namespace: namespace,
	}
}
func main() {
	log.SetFlags(log.Llongfile)
	kubeconfig := flag.String("kubeconfig", "./config", "Path to a kube config. Only required if out-of-cluster.")
	namespace := flag.String("namespace", "default", "")
	nodeName := flag.String("nodeName", "testNode", "")
	source := flag.String("source", "EventClient", "")
	eventType := flag.String("eventType", v1.EventTypeNormal, "")
	reason := flag.String("reason", "Test", "")
	messageFmt := flag.String("messageFmt", "[this is an event: %v]", "")
	message := flag.String("message", "helloworld", "")
	mode := flag.String("mode", "allinone", "")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalln(err)
	}
	eventClient := NewEventClient(config, *namespace, *nodeName)
	recorder := getEventRecorder(eventClient.client, eventClient.eventNamespace, eventClient.nodeName, *source)
	for {
		time.Sleep(time.Second)
		if *mode == "allinone" {
			recorder.Eventf(getNodeRef(*namespace, *nodeName), *eventType, *reason, *messageFmt, *message)
		}
	}
}
