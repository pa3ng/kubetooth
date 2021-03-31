package kubernetes

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/pa3ng/kubetooth/models"
	"github.com/pa3ng/kubetooth/pkg/keys"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

func (self *Client) CreateService(b models.Blockchain) error {

	for n := 0; n < b.Nodes; n++ {
		appName := fmt.Sprintf("%s-%s-%s-%d", b.Name, b.Ledger, b.Consensus, n)
		serviceName := fmt.Sprintf("%s-%d", b.Ledger, n)

		servicesClient := self.Clientset.CoreV1().Services(apiv1.NamespaceDefault)

		service := &apiv1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: serviceName,
				Labels: map[string]string{
					"name": serviceName,
				},
			},
			Spec: apiv1.ServiceSpec{
				Selector: map[string]string{
					"name": appName,
				},
				Ports: []apiv1.ServicePort{
					{
						Name:       "4004",
						Protocol:   "TCP",
						Port:       4004,
						TargetPort: intstr.FromInt(4004),
					},
					{
						Name:       "5050",
						Protocol:   "TCP",
						Port:       5050,
						TargetPort: intstr.FromInt(5050),
					},
					{
						Name:       "8008",
						Protocol:   "TCP",
						Port:       8008,
						TargetPort: intstr.FromInt(8008),
					},
					{
						Name:       "8080",
						Protocol:   "TCP",
						Port:       8080,
						TargetPort: intstr.FromInt(8080),
					},
					{
						Name:       "8800",
						Protocol:   "TCP",
						Port:       8800,
						TargetPort: intstr.FromInt(8800),
					},
				},
			},
		}

		// Create service
		fmt.Println("Creating service...")
		resultS, err := servicesClient.Create(context.TODO(), service, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		fmt.Printf("Created service %q.\n", resultS.GetObjectMeta().GetName())
	}

	return nil
}

func (self *Client) Deploy(b *models.Blockchain) error {
	deploymentsClient := self.Clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	// Setup deployment

	for n := 0; n < b.Nodes; n++ {
		appName := fmt.Sprintf("%s-%s-%s-%d", b.Name, b.Ledger, b.Consensus, n)

		deployment, err := getValidatorDeployment(appName, n, b)
		if err != nil {
			return err
		}

		// Create Deployment
		fmt.Println("Creating deployment...")
		result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
		if err != nil {
			panic(err)
		}
		fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

		if n == 0 {
			// Wait for genesis node to initiate
			fmt.Println("Loading genesis node configurations...")
			time.Sleep(20 * time.Second)
			continue
		}
		time.Sleep(5 * time.Second)
	}

	return nil
}

func (self *Client) DeployTPs(b models.Blockchain) error {
	deploymentsClient := self.Clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	servicesClient := self.Clientset.CoreV1().Services(apiv1.NamespaceDefault)

	for n := 0; n < b.Nodes; n++ {
		for _, tp := range b.Processors {
			// Setup deployment
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: tp["name"],
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": tp["name"],
						},
					},
					Template: apiv1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": tp["name"],
							},
						},
						Spec: apiv1.PodSpec{
							Containers: []apiv1.Container{
								{
									Name:  tp["name"],
									Image: tp["image"],
									Ports: []apiv1.ContainerPort{
										{
											Name:          "tcp",
											Protocol:      apiv1.ProtocolTCP,
											ContainerPort: 4004,
										},
									},
									Command: []string{"bash"},
									Args: []string{
										"-c",
										fmt.Sprintf(
											"consensource-processor -v -C tcp://sawtooth-%d:4004",
											n,
										),
									},
								},
								// containers end
							},
						},
					},
				},
			}

			// Create Deployment
			fmt.Printf("Creating deployment for tp %q...\n", tp["name"])
			result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
			if err != nil {
				panic(err)
			}
			fmt.Printf("Created tp deployment %q.\n", result.GetObjectMeta().GetName())

			// Setup service
			service := &apiv1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: tp["name"],
					Labels: map[string]string{
						"name": tp["name"],
					},
				},
				Spec: apiv1.ServiceSpec{
					Selector: map[string]string{
						"name": tp["name"],
					},
					Ports: []apiv1.ServicePort{
						{
							Name:       "4004",
							Protocol:   "TCP",
							Port:       4004,
							TargetPort: intstr.FromInt(4004),
						},
					},
				},
			}

			// Create service
			fmt.Printf("Creating service for tp %q...\n", tp["name"])
			resultS, err := servicesClient.Create(context.TODO(), service, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("Created tp service %q.\n", resultS.GetObjectMeta().GetName())
		}
	}

	return nil
}

//##################
// Configmap Helpers
//##################

func makeKeysMap(numNodes int) map[string]string {
	data := make(map[string]string)
	for i := 0; i < numNodes; i++ {
		privName := fmt.Sprintf("pbft%dpriv", i)
		pubName := fmt.Sprintf("pbft%dpub", i)
		pubKey, privKey := keys.NewKeyPair()
		data[privName] = privKey
		data[pubName] = pubKey
	}
	return data
}

//####################
// Deployment Helpers
//####################

func int32Ptr(i int32) *int32 { return &i }

//#############
// K8s Helpers
//#############

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
