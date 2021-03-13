package main

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/viper"

	"github.com/pa3ng/kubetooth/models"
	k8sclient "github.com/pa3ng/kubetooth/pkg/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newBlockchain() *models.Blockchain {
	blockchain := viper.New()
	blockchain.AddConfigPath("./config")
	blockchain.AddConfigPath(".")
	blockchain.SetConfigName("blockchain")
	err := blockchain.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %s \n", err)
	}

	var blockchainC models.Blockchain
	err = blockchain.Unmarshal(&blockchainC)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return &blockchainC
}

func main() {
	b := newBlockchain()
	fmt.Printf("Deploying %s blockchain version %s running on %s consensus\n",
		b.Ledger,
		b.Version,
		b.Consensus,
	)

	kclient, err := k8sclient.New()
	if err != nil {
		panic(err)
	}

	err = kclient.CreateKeysConfigMap("keys-config", b.Nodes)
	if err != nil {
		panic(err)
	}

	if err := Deploy(*b, kclient); err != nil {
		panic(err)
	}
}

func Deploy(b models.Blockchain, k *k8sclient.Client) error {

	clientset := k.Clientset

	// Setup Service
	appName := fmt.Sprintf("%s-%s-%s-%d", b.Name, b.Ledger, b.Consensus, 0)
	servicesClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)

	serviceName := fmt.Sprintf("%s-%d", b.Ledger, 0)
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
		panic(err)
	}
	fmt.Printf("Created service %q.\n", resultS.GetObjectMeta().GetName())

	// Setup deployment

	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	consensusEngineContainerName := fmt.Sprintf("%s-%s-engine", b.Ledger, b.Consensus)
	consensusEngineImage := fmt.Sprintf("hyperledger/%s:%s", consensusEngineContainerName, b.Version)
	validatorEntrypoint := getValidatorBashEntrypoint(0)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": appName,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": appName,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "sawtooth-validator",
							Image: "hyperledger/sawtooth-validator:chime",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "processors",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 4004,
								},
								{
									Name:          "consensus",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 5050,
								},
								{
									Name:          "validators",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 8800,
								},
							},
							EnvFrom: []apiv1.EnvFromSource{
								{
									ConfigMapRef: &apiv1.ConfigMapEnvSource{
										LocalObjectReference: apiv1.LocalObjectReference{
											Name: "keys-config",
										},
									},
								},
							},
							Command: []string{"bash"},
							Args: []string{
								"-c",
								validatorEntrypoint,
							},
						},
						// Settings TP
						{
							Name:  "sawtooth-settings-tp",
							Image: "hyperledger/sawtooth-settings-tp:chime",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "processors",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 4004,
								},
								{
									Name:          "consensus",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 5050,
								},
								{
									Name:          "validators",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 8800,
								},
							},
							Command: []string{"bash"},
							Args: []string{
								"-c",
								"settings-tp -vv -C tcp://$HOSTNAME:4004",
							},
						},
						// Consensus Engine
						{
							Name:    consensusEngineContainerName,
							Image:   consensusEngineImage,
							Command: []string{"bash"},
							Args: []string{
								"-c",
								"pbft-engine -vv --connect tcp://$HOSTNAME:5050",
							},
						},
						// Sawtooth REST API
						{
							Name:  "sawtooth-rest-api",
							Image: "hyperledger/sawtooth-rest-api:chime",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "api",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 8008,
								},
							},
							Command: []string{"bash"},
							Args: []string{
								"-c",
								"sawtooth-rest-api -vv -C tcp://$HOSTNAME:4004 -B 0.0.0.0:8008",
							},
							ReadinessProbe: &apiv1.Probe{
								Handler: apiv1.Handler{
									HTTPGet: &apiv1.HTTPGetAction{
										Path: "/status",
										Port: intstr.FromInt(8008),
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds:       10,
							},
						},
						// Sawtooth Shell
						{
							Name:    "sawtooth-shell",
							Image:   "hyperledger/sawtooth-shell:chime",
							Command: []string{"bash"},
							Args:    []string{"-c", "sawtooth keygen && tail -f /dev/null"},
						},
						// containers end
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	return nil
}

func getValidatorBashEntrypoint(node int) string {
	// TODO check which node number
	return fmt.Sprintf(vbash, node)
}

func int32Ptr(i int32) *int32 { return &i }

var vbash string = `echo "Starting validator %d" &&
if [ ! -e /etc/sawtooth/keys/validator.priv ]; then
	echo $node0priv > /etc/sawtooth/keys/validator.priv
	echo $node0pub > /etc/sawtooth/keys/validator.pub
fi &&
if [ ! -e /root/.sawtooth/keys/my_key.priv ]; then
	sawtooth keygen my_key
fi &&
if [ ! -e config-genesis.batch ]; then
	sawset genesis -k /root/.sawtooth/keys/my_key.priv -o config-genesis.batch
fi &&
sleep 30 &&
echo sawtooth.consensus.pbft.members=["\"$node0pub\",\"$node1pub\",\"$node2pub\",\"$node3pub\",\"$node4pub\""] &&
if [ ! -e config.batch ]; then
	sawset proposal create \
		-k /root/.sawtooth/keys/my_key.priv \
		sawtooth.consensus.algorithm.name=pbft \
		sawtooth.consensus.algorithm.version=1.0\
		sawtooth.consensus.pbft.members=["\"$node0pub\",\"$node1pub\",\"$node2pub\",\"$node3pub\",\"$node4pub\""] \
		sawtooth.publisher.max_batches_per_block=1200 \
		-o config.batch
fi && \
if [ ! -e /var/lib/sawtooth/genesis.batch ]; then
	sawadm genesis config-genesis.batch config.batch
fi &&
sawtooth-validator -vv \
	--endpoint tcp://$SAWTOOTH_0_SERVICE_HOST:8800 \
	--bind component:tcp://0.0.0.0:4004 \
	--bind consensus:tcp://0.0.0.0:5050 \
	--bind network:tcp://0.0.0.0:8800 \
	--scheduler parallel \
	--peering static \
	--maximum-peer-connectivity 10000
`
