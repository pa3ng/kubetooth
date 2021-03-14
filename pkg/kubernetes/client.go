package kubernetes

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/pa3ng/kubetooth/pkg/keys"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	Clientset *kubernetes.Clientset
}

func New() (*Client, error) {
	kubeconfig := getKubeconfig()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	c := &Client{
		Clientset: clientset,
	}

	return c, nil
}

func (self *Client) CreateKeysConfigMap(name string, numNodes int) error {
	configmapsClient := self.Clientset.CoreV1().ConfigMaps(apiv1.NamespaceDefault)

	cm, err := configmapsClient.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("[INFO] %s\n", err.Error())
		fmt.Println("Creating configmap...")
	} else if cm != nil {
		fmt.Printf("Found configmap %q\n", name)
		return nil
	}

	configmap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: makeKeysMap(numNodes),
	}

	resultC, err := configmapsClient.Create(context.TODO(), configmap, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("Created configmap %q\n", resultC.GetObjectMeta().GetName())

	return nil
}

func makeKeysMap(numNodes int) map[string]string {
	data := make(map[string]string)
	for i := 0; i < numNodes; i++ {
		privName := fmt.Sprintf("node%dpriv", i)
		pubName := fmt.Sprintf("node%dpub", i)
		pubKey, privKey := keys.NewKeyPair()
		data[privName] = privKey
		data[pubName] = pubKey
	}
	return data
}

func getKubeconfig() string {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String(
			"kubeconfig",
			filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file",
		)
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	return *kubeconfig
}
