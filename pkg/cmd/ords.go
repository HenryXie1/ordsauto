package cmd

import (
	"fmt"
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
	ordsdeployment         *appsv1.Deployment 
	clientset        *kubernetes.Clientset
	restConfig       *rest.Config
	rawConfig        api.Config
	ordsconfigmap    *corev1.ConfigMap
	httpconfigmap    *corev1.ConfigMap
	ordssvc          *corev1.Service
	ordsnodeportsvc  *corev1.Service
	genericclioptions.IOStreams
	UserSpecifiedOrdsname   string
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
		Use:          "ords list|create|delete [-o ordsname][-n namespace][-d dbhostname] [-p 1521] [-s dbservice] [-w syspassword] [-x apexpassword]",
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

	cmd.Flags().StringVarP(&o.UserSpecifiedApexpassword, "apexpassword", "x", "", 
	"password for apex,ords related DB schemas")
	_ = viper.BindEnv("apexpassword", "KUBECTL_PLUGINS_CURRENT_APEXPASSWORD")
	_ = viper.BindPFlag("apexpassword", cmd.Flags().Lookup("apexpassword"))	

	cmd.Flags().StringVarP(&o.UserSpecifiedNamespace, "namespace", "n", "default", 
	"namespace for ords http deployment")
	_ = viper.BindEnv("namespace", "KUBECTL_PLUGINS_CURRENT_NAMESPACE")
	_ = viper.BindPFlag("namespace", cmd.Flags().Lookup("namespace"))	

	cmd.Flags().StringVarP(&o.UserSpecifiedOrdsname, "ordsname", "o", "", 
	"name for ords http deployment")
	_ = viper.BindEnv("ordsname", "KUBECTL_PLUGINS_CURRENT_ORDSNAME")
	_ = viper.BindPFlag("ordsname", cmd.Flags().Lookup("ordsname"))	

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

	//update sys apex passwords, dbhost, db service in yaml
	  //fmt.Printf("myaml old is : %v\n",config.Ordsconfigmapyml)
		config.Ordsconfigmapyml = strings.ReplaceAll(config.Ordsconfigmapyml,"replacepwdapexordsauto",o.UserSpecifiedApexpassword)
		config.Ordsconfigmapyml = strings.ReplaceAll(config.Ordsconfigmapyml,"ordsautodbhost",o.UserSpecifiedDbhost)
		config.Ordsconfigmapyml = strings.ReplaceAll(config.Ordsconfigmapyml,"ordsautodbport",o.UserSpecifiedDbport)
		config.Ordsconfigmapyml = strings.ReplaceAll(config.Ordsconfigmapyml,"ordsautodbservice",o.UserSpecifiedService)
		config.Ordsconfigmapyml = strings.ReplaceAll(config.Ordsconfigmapyml,"replacepwdsysordsauto",o.UserSpecifiedSyspassword)
		//fmt.Printf("myaml new is : %v\n",config.Ordsconfigmapyml)
		
	//complete ords settings
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(config.Ordsyml), nil, nil)
	if err != nil {
        fmt.Printf("%#v", err)
    }
	
	o.ordsdeployment = obj.(*appsv1.Deployment)
	o.ordsdeployment.ObjectMeta.Name = o.UserSpecifiedOrdsname
	o.ordsdeployment.ObjectMeta.Namespace = o.UserSpecifiedNamespace
	//Update selector
	var ordsselector =  map[string]string {
		"ordsauto":o.UserSpecifiedOrdsname + "-DeploymentSelector",
	}
	o.ordsdeployment.Spec.Selector.MatchLabels = ordsselector
	o.ordsdeployment.Spec.Template.ObjectMeta.Labels = ordsselector
	o.ordsdeployment.Spec.Template.Spec.Volumes[0].VolumeSource.ConfigMap.LocalObjectReference = corev1.LocalObjectReference{Name: o.UserSpecifiedOrdsname + "-http-cm"}
	o.ordsdeployment.Spec.Template.Spec.Volumes[1].VolumeSource.ConfigMap.LocalObjectReference = corev1.LocalObjectReference{Name: o.UserSpecifiedOrdsname + "-ords-cm"}

	//Update LB service name
	obj, _, err = decode([]byte(config.OrdsLBsvcyml), nil, nil)
	if err != nil {
	  fmt.Printf("%#v", err)
	  }
	  o.ordssvc = obj.(*corev1.Service)
	  o.ordssvc.ObjectMeta.Name = o.UserSpecifiedOrdsname + "-svc"
	  o.ordssvc.ObjectMeta.Namespace = o.UserSpecifiedNamespace
		o.ordssvc.Spec.Selector = ordsselector
		
	//Update nodeport service name
	obj, _, err = decode([]byte(config.OrdsNodePortsvcyml), nil, nil)
	if err != nil {
	  fmt.Printf("%#v", err)
	  }
	  o.ordsnodeportsvc = obj.(*corev1.Service)
	  o.ordsnodeportsvc.ObjectMeta.Name = o.UserSpecifiedOrdsname + "-nodeport-svc"
	  o.ordsnodeportsvc.ObjectMeta.Namespace = o.UserSpecifiedNamespace
		o.ordsnodeportsvc.Spec.Selector = ordsselector

	//complete ords and http configmap settings
	obj, _, err = decode([]byte(config.Ordsconfigmapyml), nil, nil)
	if err != nil {
        fmt.Printf("%#v", err)
    }
	o.ordsconfigmap = obj.(*corev1.ConfigMap)
	o.ordsconfigmap.ObjectMeta.Name = o.UserSpecifiedOrdsname + "-ords-cm"
	o.ordsconfigmap.ObjectMeta.Namespace = o.UserSpecifiedNamespace

	obj, _, err = decode([]byte(config.Httpconfigmapyml), nil, nil)
	if err != nil {
        fmt.Printf("%#v", err)
    }
	o.httpconfigmap = obj.(*corev1.ConfigMap)
	o.httpconfigmap.ObjectMeta.Name = o.UserSpecifiedOrdsname + "-http-cm"
	o.httpconfigmap.ObjectMeta.Namespace = o.UserSpecifiedNamespace
	
	return nil
}

func (o *OrdsOperations) Validate(cmd *cobra.Command) error {
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

	if o.UserSpecifiedCreate && o.UserSpecifiedApexpassword == "" {
		_ = cmd.Usage()
		return errors.New("Must specify Apex password to create")
	}

	if o.UserSpecifiedCreate && o.UserSpecifiedOrdsname == "" {
		_ = cmd.Usage()
		return errors.New("Must specify Ords name to create")
	}

	if o.UserSpecifiedDelete && o.UserSpecifiedOrdsname == "" {
		_ = cmd.Usage()
		return errors.New("Must specify Ords name to delete")
	}


	return nil
}

func (o *OrdsOperations) Run() error {
	if o.UserSpecifiedList {
		ListOption(o)
		return nil
	}
	
	if o.UserSpecifiedCreate {
		CreateConfigmaps(o)
		CreateOrdsSchemas(o)
		CreateDeployment(o)
		CreateSvcOption(o)
	}

	if o.UserSpecifiedDelete {
		DeleteDeployment(o)
		DeleteOrdsSchemas(o)
		DeleteOrdsConfigmaps(o)
		DeleteHttpConfigmaps(o)
		DeleteSvcOption(o)
		
 	}
return nil
 
}

func ListOption(o *OrdsOperations) {
	deployclient, err := o.clientset.AppsV1().Deployments("").List(metav1.ListOptions{
		LabelSelector: "app=peordshttp",
  Limit:         100,
	})
			if err != nil {
					fmt.Println(err)
					return
	}
if 	len(deployclient.Items) == 0 {
	fmt.Printf("Didn't found Ords Deployment with label app=peordshttp \n")
	 
} else {
for i := 0;i < len(deployclient.Items);i++ {
	fmt.Printf("Found %v Deployment with label app=peordshttp in namespace %v\n", deployclient.Items[i].ObjectMeta.Name,deployclient.Items[i].ObjectMeta.Namespace)
	 }
   }

   CreateSqlplusPod(o)
   fmt.Printf("List Apex Details in Target DB....\n")
	sqltext := "sqlplus " + "sys/" + o.UserSpecifiedSyspassword + "@" + o.UserSpecifiedDbhost + ":" + o.UserSpecifiedDbport + "/" + o.UserSpecifiedService + " as sysdba " + "@apexrelease.sql"
	//fmt.Println(sqltext)
	SqlCommand := []string{"/bin/sh", "-c", sqltext}	 
	Podname := "sqlpluspod"
	err = ExecPodCmd(o,Podname,SqlCommand)
	if err != nil {
		fmt.Printf("Error occured in the Pod ,Sqlcommand %q. Error: %+v\n", SqlCommand, err)
	} 

	fmt.Printf("List ORDS Details in Target DB....\n")
	sqltext = "sqlplus " + "sys/" + o.UserSpecifiedSyspassword + "@" + o.UserSpecifiedDbhost + ":" + o.UserSpecifiedDbport + "/" + o.UserSpecifiedService + " as sysdba " + "@ordsversion.sql"
	//fmt.Println(sqltext)
	SqlCommand = []string{"/bin/sh", "-c", sqltext}	 
	Podname = "sqlpluspod"
	err = ExecPodCmd(o,Podname,SqlCommand)
	if err != nil {
		fmt.Printf("Error occured in the Pod ,Sqlcommand %q. Error: %+v\n", SqlCommand, err)
	} 
	DeleteSqlplusPod(o)
}

func CreateConfigmaps(o *OrdsOperations){
	fmt.Printf("Creating configmap %v in namespace %v...\n",o.ordsconfigmap.ObjectMeta.Name,o.ordsconfigmap.ObjectMeta.Namespace)
	ConfigMapclient := o.clientset.CoreV1().ConfigMaps(o.UserSpecifiedNamespace)
    result, err := ConfigMapclient.Create(o.ordsconfigmap)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Created configmap %q.\n", result.GetObjectMeta().GetName())

	fmt.Printf("Creating configmap %v in namespace %v...\n",o.httpconfigmap.ObjectMeta.Name,o.httpconfigmap.ObjectMeta.Namespace)
	ConfigMapclient = o.clientset.CoreV1().ConfigMaps(o.UserSpecifiedNamespace)
    result, err = ConfigMapclient.Create(o.httpconfigmap)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Created configmap %q.\n", result.GetObjectMeta().GetName())
	time.Sleep(5 * time.Second)
	
}

func DeleteOrdsConfigmaps(o *OrdsOperations){
	fmt.Printf("Deleting Ords configmap %q with label app=peordshttp in namespace %v...\n",o.UserSpecifiedOrdsname + "-ords-cm",o.UserSpecifiedNamespace)
	ConfigMapclient := o.clientset.CoreV1().ConfigMaps(o.UserSpecifiedNamespace)
	deletePolicy := metav1.DeletePropagationForeground
	listOptions := metav1.ListOptions{
		LabelSelector: "app=peordshttp",
		FieldSelector: "metadata.name=" + o.UserSpecifiedOrdsname + "-ords-cm",
        Limit:         100,
	}
	list, err := ConfigMapclient.List(listOptions)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Printf("length of list: %v",len(list.Items))
	if len(list.Items) == 0 {
		fmt.Println("No configmap found\n")
		return
	} else {
	for _, d := range list.Items {
		fmt.Printf(" * %s \n", d.Name)
	  }
    }
    if err := ConfigMapclient.DeleteCollection(&metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	    },listOptions); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Deleted Ords configmaps in namespace %v.\n",o.UserSpecifiedNamespace)
	time.Sleep(5 * time.Second)
	
}

func DeleteHttpConfigmaps(o *OrdsOperations){
	fmt.Printf("Deleting Http configmap %q with label app=peordshttp in namespace %v...\n",o.UserSpecifiedOrdsname + "-http-cm",o.UserSpecifiedNamespace)
	ConfigMapclient := o.clientset.CoreV1().ConfigMaps(o.UserSpecifiedNamespace)
	deletePolicy := metav1.DeletePropagationForeground
	listOptions := metav1.ListOptions{
		LabelSelector: "app=peordshttp",
		FieldSelector: "metadata.name=" + o.UserSpecifiedOrdsname + "-http-cm",
        Limit:         100,
	}
	list, err := ConfigMapclient.List(listOptions)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Printf("length of list: %v",len(list.Items))
	if len(list.Items) == 0 {
		fmt.Println("No configmap found\n")
		return
	} else {
	for _, d := range list.Items {
		fmt.Printf(" * %s \n", d.Name)
	  }
    }
    if err := ConfigMapclient.DeleteCollection(&metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	    },listOptions); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Deleted Http configmaps in namespace %v.\n",o.UserSpecifiedNamespace)
	time.Sleep(5 * time.Second)
	
}

func CreateDeployment(o *OrdsOperations) {
	fmt.Printf("Creating Ords Deployment %q in namespace %v...\n",o.UserSpecifiedOrdsname, o.UserSpecifiedNamespace)
	Deployclient := o.clientset.AppsV1().Deployments(o.UserSpecifiedNamespace)
		result, err := Deployclient.Create(o.ordsdeployment)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Created Ords Deployment: %q.\n", result.GetObjectMeta().GetName())
	fmt.Println("Please wait about 3-10 min for Pod to be fully up,use kubectl get po to check.")
	time.Sleep(5 * time.Second)
	
}

func CreateOrdsSchemas(o *OrdsOperations) {
	CreateOrdsPod(o)

	fmt.Printf("Create Ords schemas(ORDS_METADA,ORDS_PUBLIC_USER) in Target DB Host %v....\n",o.UserSpecifiedDbhost)
	ordstext := "mv /opt/oracle/ords/config/ords/defaults.xml /tmp;cp /mnt/k8s/ords_params.properties /tmp/ords_params.properties;java -jar /opt/oracle/ords/ords.war install --parameterFile /tmp/ords_params.properties simple"
	OrdsCommand := []string{"/bin/sh", "-c", ordstext}	 
	Podname := "ordspod"
	err := ExecPodCmd(o,Podname,OrdsCommand)
	if err != nil {
		fmt.Printf("Error occured in the Pod ,OrdsCommand %q. Error: %+v\n", OrdsCommand, err)
	} 

	DeleteOrdsPod(o)
	
}

func DeleteDeployment(o *OrdsOperations) {
	fmt.Printf("Deleting Ords Deployment %q in namespace %v...\n",o.UserSpecifiedOrdsname,o.UserSpecifiedNamespace)
	Deployclient := o.clientset.AppsV1().Deployments(o.UserSpecifiedNamespace)
	deletePolicy := metav1.DeletePropagationForeground
	listOptions := metav1.ListOptions{
		LabelSelector: "app=peordshttp",
		FieldSelector: "metadata.name=" + o.UserSpecifiedOrdsname,
        Limit:         100,
	}
	list, err := Deployclient.List(listOptions)
	if err != nil {
		fmt.Println(err)
		return
	}
	
	if len(list.Items) == 0 {
		fmt.Printf("No %q with label app=peordshttp Deployment found\n",o.UserSpecifiedOrdsname)
		return
	} else {
	for _, d := range list.Items {
		fmt.Printf(" * %s \n", d.Name)
	  }
    }
    if err := Deployclient.DeleteCollection(&metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	    },listOptions); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Deleted Ords %q deployment in namespace %v.\n",o.UserSpecifiedOrdsname, o.UserSpecifiedNamespace)
	time.Sleep(5 * time.Second)
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
		Namespace(o.UserSpecifiedNamespace).
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
		Namespace:    o.UserSpecifiedNamespace,
	}
	podSpecs := corev1.PodSpec{
		//ImagePullSecrets: []corev1.LocalObjectReference{{
		//	Name: "iad-ocir-secret",
		//}},
		Containers:    []corev1.Container{{
			Name: "sqlpluspod",
			Image: "iad.ocir.io/espsnonprodint/autostg/instantclient-apex19:v1",
		}},
	}
	pod := corev1.Pod{
			TypeMeta:   typeMetadata,
			ObjectMeta: objectMetadata,
			Spec:       podSpecs,
}
fmt.Println("Creating sqlpluspod .......")
createdPod, err := o.clientset.CoreV1().Pods(o.UserSpecifiedNamespace).Create(&pod)
if err != nil {
	return fmt.Errorf("error in creating sqlpluspod: %v", err)
}
time.Sleep(5 * time.Second)
verifyPodState := func() bool {
	podStatus, err := o.clientset.CoreV1().Pods(o.UserSpecifiedNamespace).Get(createdPod.Name, metav1.GetOptions{})
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

func CreateOrdsPod(o *OrdsOperations) error{

	typeMetadata := metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
	}
	objectMetadata := metav1.ObjectMeta{
		Name: "ordspod",
		Namespace:    o.UserSpecifiedNamespace,
	}

	configmapvolume := &corev1.ConfigMapVolumeSource{
		LocalObjectReference: corev1.LocalObjectReference{Name: o.UserSpecifiedOrdsname + "-ords-cm"},
	}

	podSpecs := corev1.PodSpec{
		//ImagePullSecrets: []corev1.LocalObjectReference{{
		//	Name: "iad-ocir-secret",
		//}},
		Volumes:  []corev1.Volume{{
			Name: "ords-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: configmapvolume,
			},
		}},
		Containers:    []corev1.Container{{
			Name: "ordspod",
			Image: "iad.ocir.io/espsnonprodint/autostg/apexords:v19",
			VolumeMounts: []corev1.VolumeMount{{
				Name: "ords-config",
				MountPath: "/mnt/k8s",
			}},
		}},
	}
	pod := corev1.Pod{
			TypeMeta:   typeMetadata,
			ObjectMeta: objectMetadata,
			Spec:       podSpecs,
}
fmt.Println("Creating ords pod .......")
createdPod, err := o.clientset.CoreV1().Pods(o.UserSpecifiedNamespace).Create(&pod)
if err != nil {
	return fmt.Errorf("error in creating ords pod: %v", err)
}
time.Sleep(5 * time.Second)
verifyPodState := func() bool {
	podStatus, err := o.clientset.CoreV1().Pods(o.UserSpecifiedNamespace).Get(createdPod.Name, metav1.GetOptions{})
	if err != nil {
		return false
	} 
	
	if podStatus.Status.Phase == corev1.PodRunning {
		return true
	} 
	return false
}
//10 min timeout for starting pod
for i:=0;i<120;i++{
	if  !verifyPodState() { 
		fmt.Println("waiting for ordspod to start.......")
		time.Sleep(5 * time.Second)
		
	} else {
		fmt.Println("ords pod started.......")
		return nil
	}
}
return fmt.Errorf("Timeout to start ordspod : %v", err)

}

func DeleteSqlplusPod(o *OrdsOperations) error {

fmt.Println("Deleting sqlpluspod .......")
deletePolicy := metav1.DeletePropagationForeground

err := o.clientset.CoreV1().Pods(o.UserSpecifiedNamespace).Delete("sqlpluspod", 
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

func DeleteOrdsPod(o *OrdsOperations) error {

	fmt.Println("Deleting ordspod .......")
	deletePolicy := metav1.DeletePropagationForeground
	
	err := o.clientset.CoreV1().Pods(o.UserSpecifiedNamespace).Delete("ordspod", 
			&metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
			})
	if err != nil {
	return fmt.Errorf("error in deleting ordspod: %v", err)
	} else {
	time.Sleep(5 * time.Second)
	fmt.Println("Deleted ordspod .......")
	return nil
	}
	
	}

func CreateSvcOption(o *OrdsOperations) {
	fmt.Printf("Creating nodeport service for Ords in namespace %v...\n",o.UserSpecifiedNamespace)
	Svcclient := o.clientset.CoreV1().Services(o.UserSpecifiedNamespace)
    result, err := Svcclient.Create(o.ordsnodeportsvc)
	if err != nil {
		fmt.Println(err)
		return 
	}
	fmt.Printf("Created service %q.\n\n", result.GetObjectMeta().GetName())
	NodePortResult := result.Spec.Ports[0].NodePort
	//find a host IP address for nodeport service connections
	NodeStatus, err := o.clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		fmt.Println(err)
		return 
	} 
	OrdsHostip := NodeStatus.Items[0].Status.Addresses[0].Address


	fmt.Printf("Creating Load Balancer service for Ords in namespace %v...\n",o.UserSpecifiedNamespace)
	Svcclient = o.clientset.CoreV1().Services(o.UserSpecifiedNamespace)
    result, err = Svcclient.Create(o.ordssvc)
	if err != nil {
		fmt.Println(err)
		return 
	}
	time.Sleep(5 * time.Second)
	verifySvcState := func() string {
		SvcStatus, err := o.clientset.CoreV1().Services(o.UserSpecifiedNamespace).Get(o.UserSpecifiedOrdsname + "-svc", metav1.GetOptions{})
		if err != nil {
			fmt.Println(err)
			return ""
		} 
		//fmt.Printf("external ip: %v",SvcStatus.Status.LoadBalancer.Ingress)
		if len(SvcStatus.Status.LoadBalancer.Ingress) != 0 {
		   	return SvcStatus.Status.LoadBalancer.Ingress[0].IP
		} else { 
				return ""
		}
 }
	//3 min timeout for getting external IP
   var OrdsExternalIP string
  for i:=0;i<36;i++{
	OrdsExternalIP = verifySvcState()
		if  OrdsExternalIP == "" { 
			fmt.Println("waiting for LoadBalancer External IP .......")
			time.Sleep(5 * time.Second)
			
		} else {
			fmt.Printf("LoadBalancer External IP is %v\n",OrdsExternalIP)
			break
		}
	}
	fmt.Printf("Created service %q.\n", result.GetObjectMeta().GetName())
	fmt.Printf("Url to access Apex service via nodeport: http://%v:%v \n",OrdsHostip ,NodePortResult)
	fmt.Printf("Url to access Apex service via Loadbalancer: http://%v\n",OrdsExternalIP)
	fmt.Println("workspace:internal,username:admin,password:Welcome1` (Use apxchpwd.sql to change it)" )
	fmt.Println("If Apex runtime only is installed,internal workspace is not available." )
	time.Sleep(5 * time.Second)
}


func DeleteSvcOption(o *OrdsOperations) {
	fmt.Printf("Deleting Load Balancer service %v with label app=peordsauto in namespace %v...\n",o.UserSpecifiedOrdsname + "-svc",o.UserSpecifiedNamespace)
	  Svcclient := o.clientset.CoreV1().Services(o.UserSpecifiedNamespace)
	  deletePolicy := metav1.DeletePropagationForeground
	  listOptions := metav1.ListOptions{
				  LabelSelector: "app=peordsauto",
				  FieldSelector: "metadata.name=" + o.UserSpecifiedOrdsname + "-svc",
		  Limit:         100,
	  }
	  list, err := Svcclient.List(listOptions)
	  if err != nil {
		  fmt.Println(err)
		   
	  }
	  
	  if len(list.Items) == 0 {
		  fmt.Println("No Services found")
		   
	  } else {
	  for _, d := range list.Items {
		  fmt.Printf(" * %s \n", d.Name)
		}
	  }
	  if err := Svcclient.Delete(o.UserSpecifiedOrdsname + "-svc", &metav1.DeleteOptions{
		  PropagationPolicy: &deletePolicy,
		  }); err != nil {
				  fmt.Println(err)
				   
	  }
		fmt.Printf("Deleted load balancer services in namespace %v.\n",o.UserSpecifiedNamespace)
		
	  fmt.Printf("Deleting nodeport service %v with label app=peordsauto in namespace %v...\n",o.UserSpecifiedOrdsname + "-svc",o.UserSpecifiedNamespace)
	  Svcclient = o.clientset.CoreV1().Services(o.UserSpecifiedNamespace)
	  deletePolicy = metav1.DeletePropagationForeground
	  listOptions = metav1.ListOptions{
				  LabelSelector: "app=peordsauto",
				  FieldSelector: "metadata.name=" + o.UserSpecifiedOrdsname + "-nodeport-svc",
		  Limit:         100,
	  }
	  list, err = Svcclient.List(listOptions)
	  if err != nil {
		  fmt.Println(err)
		   
	  }
	  
	  if len(list.Items) == 0 {
		  fmt.Println("No Services found")
		   
	  } else {
	  for _, d := range list.Items {
		  fmt.Printf(" * %s \n", d.Name)
		}
	  }
	  if err := Svcclient.Delete(o.UserSpecifiedOrdsname + "-nodeport-svc", &metav1.DeleteOptions{
		  PropagationPolicy: &deletePolicy,
		  }); err != nil {
				  fmt.Println(err)
				   
	  }
	  fmt.Printf("Deleted nodeport services in namespace %v.\n",o.UserSpecifiedNamespace)
	  time.Sleep(5 * time.Second)
	  return 
  
  }