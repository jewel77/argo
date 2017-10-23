package client

import (
	"fmt"
	"time"

	wfv1 "github.com/argoproj/argo/api/workflow/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func CreateCustomResourceDefinition(clientset apiextensionsclient.Interface) (*apiextensionsv1beta1.CustomResourceDefinition, error) {
	fmt.Printf("Creating Workflow CRD\n")
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: wfv1.CRDFullName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   wfv1.CRDGroup,
			Version: wfv1.SchemeGroupVersion.Version,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:     wfv1.CRDPlural,
				Kind:       wfv1.CRDKind,
				ShortNames: []string{wfv1.CRDShortName},
			},
		},
	}

	_, err := clientset.Apiextensions().CustomResourceDefinitions().Create(crd)
	if err != nil {
		return nil, err
	}

	// wait for CRD being established
	err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		crd, err = clientset.Apiextensions().CustomResourceDefinitions().Get(wfv1.CRDFullName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1beta1.Established:
				if cond.Status == apiextensionsv1beta1.ConditionTrue {
					return true, err
				}
			case apiextensionsv1beta1.NamesAccepted:
				if cond.Status == apiextensionsv1beta1.ConditionFalse {
					fmt.Printf("Name conflict: %v\n", cond.Reason)
				}
			}
		}
		return false, err
	})
	if err != nil {
		deleteErr := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(wfv1.CRDFullName, nil)
		if deleteErr != nil {
			return nil, errors.NewAggregate([]error{err, deleteErr})
		}
		return nil, err
	}
	return crd, nil
}

// func WaitWorkflowCompletion(wfClient *rest.RESTClient, name string) error {
// 	return wait.Poll(100*time.Millisecond, 10*time.Second, func() (bool, error) {
// 		var wf wfv1.Workflow
// 		err := wfClient.Get().
// 			Resource(wfv1.CRDPlural).
// 			Namespace(apiv1.NamespaceDefault).
// 			Name(name).
// 			Do().Into(&wf)

// 		if err == nil && wf.Status == wfv1.WorkflowStatusFailed {
// 			return true, nil
// 		}

// 		return false, err
// 	})
// }