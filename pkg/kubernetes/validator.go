package kubernetes

import (
	"fmt"

	"github.com/pa3ng/kubetooth/models"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getValidatorDeployment(appName string, n int, b *models.Blockchain) (*appsv1.Deployment, error) {
	consensusEngineContainerName := fmt.Sprintf("%s-%s-engine", b.Ledger, b.Consensus)
	consensusEngineImage := fmt.Sprintf("hyperledger/%s:%s", consensusEngineContainerName, b.Version)

	validatorEntrypoint, err := getValidatorBashEntrypoint(n, b.Nodes)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return &appsv1.Deployment{
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
								// ReadinessProbe: &apiv1.Probe{
								// 	Handler: apiv1.Handler{
								// 		HTTPGet: &apiv1.HTTPGetAction{
								// 			Path: "/status",
								// 			Port: intstr.FromInt(8008),
								// 		},
								// 	},
								// 	InitialDelaySeconds: 15,
								// 	PeriodSeconds:       10,
								// },
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
		}, nil
	}

	envSection := []apiv1.EnvVar{
		{
			Name: fmt.Sprintf("%s%dpriv", b.Consensus, n),
			ValueFrom: &apiv1.EnvVarSource{
				ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
					Key: fmt.Sprintf("%s%dpriv", b.Consensus, n),
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: "keys-config",
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("%s%dpub", b.Consensus, n),
			ValueFrom: &apiv1.EnvVarSource{
				ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
					Key: fmt.Sprintf("%s%dpub", b.Consensus, n),
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: "keys-config",
					},
				},
			},
		},
	}

	return &appsv1.Deployment{
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
							Env:     envSection,
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
							// ReadinessProbe: &apiv1.Probe{
							// 	Handler: apiv1.Handler{
							// 		HTTPGet: &apiv1.HTTPGetAction{
							// 			Path: "/status",
							// 			Port: intstr.FromInt(8008),
							// 		},
							// 	},
							// 	InitialDelaySeconds: 15,
							// 	PeriodSeconds:       10,
							// },
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
	}, nil

}

func getValidatorBashEntrypoint(nodeNum, numNodes int) (string, error) {
	if nodeNum < 0 {
		return "", fmt.Errorf("Node number must be indexed 0 or higher")
	}

	entrybash := fmt.Sprintf(validatorEntryBash, nodeNum, nodeNum, nodeNum)

	var validatorBash = fmt.Sprintf(validatorBinaryCallFormat, nodeNum)
	if nodeNum == 0 {
		memberNames := ""
		for i := 0; i < numNodes; i++ {
			name := fmt.Sprintf(`\"$pbft%dpub\"`, i)
			memberNames = memberNames + name
			if i < numNodes-1 {
				memberNames = memberNames + ","
			}
		}
		sawtoothPbftMembersList := fmt.Sprintf(`sawtooth.consensus.pbft.members=["%s"]`, memberNames)
		genesisNodeConfigBash := fmt.Sprintf(
			genesisNodePbftConfigBashFormat,
			sawtoothPbftMembersList,
			sawtoothPbftMembersList,
		)
		return entrybash + genesisNodeConfigBash + validatorBash, nil
	} else {
		keygenBash := `sawtooth keygen my_key &&`
		validatorBash = keygenBash + validatorBash
		for i := 0; i < numNodes; i++ {
			if i == nodeNum {
				break
			}
			peerArg := fmt.Sprintf(`--peers tcp://$SAWTOOTH_%d_SERVICE_HOST:8800`, i)
			validatorBash = validatorBash + " \\\n" + peerArg
		}
	}
	return entrybash + validatorBash, nil
}

var genesisNodePbftConfigBashFormat string = `
if [ ! -e /root/.sawtooth/keys/my_key.priv ]; then
	sawtooth keygen my_key
fi &&
if [ ! -e config-genesis.batch ]; then
	sawset genesis -k /root/.sawtooth/keys/my_key.priv -o config-genesis.batch
fi &&
sleep 30 &&
echo %s &&
if [ ! -e config.batch ]; then
	sawset proposal create \
			-k /root/.sawtooth/keys/my_key.priv \
			sawtooth.consensus.algorithm.name=pbft \
			sawtooth.consensus.algorithm.version=1.0\
			%s \
			sawtooth.publisher.max_batches_per_block=1200 \
			-o config.batch
fi && \
if [ ! -s /var/lib/sawtooth/backup/backupdata ]; then
	sawadm genesis config-genesis.batch config.batch;
else
	sawadm blockstore restore /var/lib/sawtooth/backup/backupdata
fi &&`

var validatorBinaryCallFormat string = `
sawtooth-validator -vv \
	--endpoint tcp://$SAWTOOTH_%d_SERVICE_HOST:8800 \
	--bind component:tcp://0.0.0.0:4004 \
	--bind consensus:tcp://0.0.0.0:5050 \
	--bind network:tcp://0.0.0.0:8800 \
	--scheduler parallel \
	--peering static \
	--maximum-peer-connectivity 10000`

var validatorEntryBash string = `echo "Starting validator %d" &&
if [ ! -e /etc/sawtooth/keys/validator.priv ]; then
	echo $pbft%dpriv > /etc/sawtooth/keys/validator.priv
	echo $pbft%dpub > /etc/sawtooth/keys/validator.pub
fi &&
`
