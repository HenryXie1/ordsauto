package cmd

import (
	"fmt"
	//"io"
	//"bytes"
	//"io/ioutil"
	"github.com/pkg/errors"
	//log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	//utilexec "k8s.io/client-go/util/exec"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	//"k8s.io/api/extensions/v1beta1"
	"os"
	//"os/exec"
	//"path/filepath"
	"strings"
	"time"
	"ordsauto/pkg/config"
)

type OrdsOperations struct {
	configFlags      *genericclioptions.ConfigFlags
	ordshttp         *appsv1.Deployment 
	clientset        *kubernetes.Clientset
	restConfig       *rest.Config
	rawConfig        api.Config
	genericclioptions.IOStreams
	UserSpecifiedNamespace  string
	UserSpecifiedDbhost     string
	UserSpecifiedDbport     string
	UserSpecifiedService    string
	UserSpecifiedSyspassword   string
	UserSpecifiedApexpassword   string
	UserSpecifiedCreate   bool
	UserSpecifiedDelete   bool
	UserSpecifiedList     bool
	

}

// NewOrdsOperations provides an instance of OrdsOperations with default values
func NewOrdsOperations(streams genericclioptions.IOStreams) *OrdsOperations {
	return &OrdsOperations{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams: streams,
	}
}

// NewCmdOrds provides a cobra command wrapping OrdsOperations
func NewCmdOrds(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewOrdsOperations(streams)

	cmd := &cobra.Command{
		Use:          "ords list|create|delete [-n namespace][-d dbhostname] [-p 1521] [-s dbservice] [-w syspassword] [-x apexpassword] ",
		Short:        "create or delete ords + http deployment in K8S",
		Example:      fmt.Sprintf(config.OrdsExample),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			  
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(c); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&o.UserSpecifiedDbhost, "dbhost", "d", "", "DB hostname or IP address")
	_ = viper.BindEnv("dbhost", "KUBECTL_PLUGINS_CURRENT_DBHOST")
	_ = viper.BindPFlag("dbhost", cmd.Flags().Lookup("dbhost"))

	cmd.Flags().StringVarP(&o.UserSpecifiedDbport, "dbport", "p", "1521", "DB port to access")
	_ = viper.BindEnv("dbport", "KUBECTL_PLUGINS_CURRENT_dbport")
	_ = viper.BindPFlag("dbport", cmd.Flags().Lookup("dbport"))

	cmd.Flags().StringVarP(&o.UserSpecifiedService, "service", "s", "", "DB service to access")
	_ = viper.BindEnv("service", "KUBECTL_PLUGINS_CURRENT_SERVICE")
	_ = viper.BindPFlag("service", cmd.Flags().Lookup("service"))

	cmd.Flags().StringVarP(&o.UserSpecifiedSyspassword, "syspassword", "w", "",
	"sys password of DB service")
	_ = viper.BindEnv("syspassword", "KUBECTL_PLUGINS_CURRENT_SYSPASSWORD")
	_ = viper.BindPFlag("syspassword", cmd.Flags().Lookup("syspassword"))

	cmd.Flags().StringVarP(&o.UserSpecifiedApexpassword, "apexpassword", "x", "password", 
	"password for apex related DB schemas")
	_ = viper.BindEnv("apexpassword", "KUBECTL_PLUGINS_CURRENT_APEXPASSWORD")
	_ = viper.BindPFlag("apexpassword", cmd.Flags().Lookup("apexpassword"))	

	cmd.Flags().StringVarP(&o.UserSpecifiedNamespace, "namespace", "n", "default", 
	"namespace for ords http deployment")
	_ = viper.BindEnv("namespace", "KUBECTL_PLUGINS_CURRENT_NAMESPACE")
	_ = viper.BindPFlag("namespace", cmd.Flags().Lookup("namespace"))	

	return cmd
}

func (o *OrdsOperations) Complete(cmd *cobra.Command, args []string) error {
	
	if len(args) != 1 {
		_ = cmd.Usage()
		return errors.New("Please check kubectl-ords -h for usage")
	}

	switch strings.ToUpper(args[0]) {
	case "CREATE":
		o.UserSpecifiedCreate = true
	case "DELETE":
		o.UserSpecifiedDelete = true
	case "LIST":
		o.UserSpecifiedList = true
	default:
		_ = cmd.Usage()
		return errors.New("Please check kubectl-ords -h for usage")
	}

	var err error
	o.rawConfig, err = o.configFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	o.restConfig, err = o.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	o.restConfig.Timeout = 180 * time.Second
	o.clientset, err = kubernetes.NewForConfig(o.restConfig)
	if err != nil {
		return err
	}
		
	//complete ordshttp settings
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(config.OrdsHttpyml), nil, nil)
	if err != nil {
        fmt.Printf("%#v", err)
    }
	obj, _, err = decode([]byte(config.OrdsHttpyml), nil, nil)
	if err != nil {
        fmt.Printf("%#v", err)
    }
	o.ordshttp = obj.(*appsv1.Deployment)
	
	o.ordshttp.ObjectMeta.Namespace = o.UserSpecifiedNamespace
	
	return nil
}

func (o *OrdsOperations) Validate(cmd *cobra.Command) error {
	if o.UserSpecifiedList {
		deployclient, err := o.clientset.AppsV1().Deployments("").List(metav1.ListOptions{
			LabelSelector: "app=peordshttp",
      Limit:         100,
		})
				if err != nil {
						panic(err.Error())
		}
	for i := 0;i < len(deployclient.Items);i++ {
		fmt.Printf("Found %v Deployment with label app=peordshttp in namespace %v\n", deployclient.Items[i].ObjectMeta.Name,deployclient.Items[i].ObjectMeta.Namespace)
		 }
	return nil
}
	
	if o.UserSpecifiedDbhost == "" {
		_ = cmd.Usage()
		return errors.New("Must specify DB hostname name")
	}

	if o.UserSpecifiedService == "" {
		_ = cmd.Usage()
		return errors.New("Must specify DB Service")
	}
   
	if o.UserSpecifiedSyspassword == "" {
		_ = cmd.Usage()
		return errors.New("Must specify sys password of DB Service")
	}

	return nil
}

func (o *OrdsOperations) Run() error {
	
	if o.UserSpecifiedCreate {
		CreateDeployment(o)
	}

	if o.UserSpecifiedDelete {
		DeleteDeployment(o)
		DeleteOrdsSchemas(o)
		
 	}
return nil
 
}

func CreateDeployment(o *OrdsOperations) {
	fmt.Printf("Creating Ords Http Deployment in namespace %v...\n",o.UserSpecifiedNamespace)
	Deployclient := o.clientset.AppsV1().Deployments(o.UserSpecifiedNamespace)
		result, err := Deployclient.Create(o.ordshttp)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created Ords Http Deployment: %q.\n", result.GetObjectMeta().GetName())
	
}

func DeleteDeployment(o *OrdsOperations) {
	fmt.Printf("Deleting Ords Http Deployment in namespace %v...\n",o.UserSpecifiedNamespace)
	Deployclient := o.clientset.AppsV1().Deployments(o.UserSpecifiedNamespace)
	deletePolicy := metav1.DeletePropagationForeground
	listOptions := metav1.ListOptions{
        LabelSelector: "app=peordshttp",
        Limit:         100,
	}
	list, err := Deployclient.List(listOptions)
	if err != nil {
		panic(err)
	}
	
	if len(list.Items) == 0 {
		fmt.Println("No ords-http with label app=peordshttp Deployment found")
		return
	} else {
	for _, d := range list.Items {
		fmt.Printf(" * %s \n", d.Name)
	  }
    }
    if err := Deployclient.DeleteCollection(&metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	    },listOptions); err != nil {
		panic(err)
	}
	fmt.Printf("Deleted ords-http deployment in namespace %v.\n",o.UserSpecifiedNamespace)
}

func DeleteOrdsSchemas(o *OrdsOperations) {
	CreateSqlplusPod(o)

	fmt.Printf("Dropping Ords schemas(ORDS_METADA,ORDS_PUBLIC_USER) in Target DB....\n")
	sqltext := "sqlplus " + "sys/" + o.UserSpecifiedSyspassword + "@" + o.UserSpecifiedDbhost + ":" + o.UserSpecifiedDbport + "/" + o.UserSpecifiedService + " as sysdba " + "@dropords.sql "
	//fmt.Println(sqltext)
	SqlCommand := []string{"/bin/sh", "-c", sqltext}	 
	Podname := "sqlpluspod"
	err := ExecPodCmd(o,Podname,SqlCommand)
	if err != nil {
		fmt.Printf("Error occured in the Pod ,Sqlcommand %q. Error: %+v\n", SqlCommand, err)
	} 

	DeleteSqlplusPod(o)

}

func ExecPodCmd(o *OrdsOperations,Podname string,SqlCommand []string) error {
	
	execReq := o.clientset.CoreV1().RESTClient().Post().
	    Resource("pods").
		Name(Podname).
		Namespace("default").
		SubResource("exec")

    execReq.VersionedParams(&corev1.PodExecOptions{
		Command:   SqlCommand,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(o.restConfig, "POST", execReq.URL())
	if err != nil {
		return fmt.Errorf("error while creating Executor: %v", err)
	}

	err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Tty:    false,
		})
	if err != nil {
		return fmt.Errorf("error in Stream: %v", err)
	} else {
		return nil
	}
	
}

func CreateSqlplusPod(o *OrdsOperations) error{

	typeMetadata := metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
	}
	objectMetadata := metav1.ObjectMeta{
		Name: "sqlpluspod",
		Namespace:    "default",
	}
	podSpecs := corev1.PodSpec{
		Containers:    []corev1.Container{{
			Name: "sqlpluspod",
			Image: "iad.ocir.io/espsnonprodint/livesqlsandbox/instantclient:apex19",
		}},
	}
	pod := corev1.Pod{
			TypeMeta:   typeMetadata,
			ObjectMeta: objectMetadata,
			Spec:       podSpecs,
}
fmt.Println("Creating sqlpluspod .......")
createdPod, err := o.clientset.CoreV1().Pods("default").Create(&pod)
if err != nil {
	return fmt.Errorf("error in creating sqlpluspod: %v", err)
}
time.Sleep(5 * time.Second)
verifyPodState := func() bool {
	podStatus, err := o.clientset.CoreV1().Pods("default").Get(createdPod.Name, metav1.GetOptions{})
	if err != nil {
		return false
	} 
	
	if podStatus.Status.Phase == corev1.PodRunning {
		return true
	} 
	return false
}
//3 min timeout for starting pod
for i:=0;i<36;i++{
	if  !verifyPodState() { 
		fmt.Println("waiting for sqlpluspod to start.......")
		time.Sleep(5 * time.Second)
		
	} else {
		fmt.Println("sqlpluspod started.......")
		return nil
	}
}
return fmt.Errorf("Timeout to start sqlpluspod : %v", err)


}

func DeleteSqlplusPod(o *OrdsOperations) error {

fmt.Println("Deleting sqlpluspod .......")
deletePolicy := metav1.DeletePropagationForeground

err := o.clientset.CoreV1().Pods("default").Delete("sqlpluspod", 
		&metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
		})
if err != nil {
return fmt.Errorf("error in deleting sqlpluspod: %v", err)
} else {
time.Sleep(5 * time.Second)
fmt.Println("Deleted sqlpluspod .......")
return nil
}

}