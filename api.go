package main

import (
	"log"
	"os"
	"time"

	"github.com/lean-soft/play-with-docker/config"
	"github.com/lean-soft/play-with-docker/docker"
	"github.com/lean-soft/play-with-docker/event"
	"github.com/lean-soft/play-with-docker/handlers"
	"github.com/lean-soft/play-with-docker/id"
	"github.com/lean-soft/play-with-docker/k8s"
	"github.com/lean-soft/play-with-docker/provisioner"
	"github.com/lean-soft/play-with-docker/pwd"
	"github.com/lean-soft/play-with-docker/pwd/types"
	"github.com/lean-soft/play-with-docker/scheduler"
	"github.com/lean-soft/play-with-docker/scheduler/task"
	"github.com/lean-soft/play-with-docker/storage"
)

func main() {

	config.ParseFlags()

	//todo: allow reading configuration from file or env-var
	var domain = os.Getenv("PLAYGROUND_DOMAIN")
	var image = os.Getenv("DEFAULT_DIND_IMAGE")
	var assetsDir = os.Getenv("ASSETS_DIR")
	var defaultSessionDuration = os.Getenv("DefaultSessionDuration")

	// update this before docker build and fill in actual target domain name
	// e.g. play-with-docker.cn
	//		play-with-k8s.cn
	// for local develoment, leave this blank or use localhost
	config.PlaygroundDomain = domain

	// update this for the DinD image name
	// e.g. pwd use franela/dind
	//		pwk use franela/k8s
	config.DefaultDinDImage = image

	// set default session duration
	config.DefaultSessionDuration = defaultSessionDuration

	log.Println("Env PLAYGROUND_DOMAIN=" + domain)
	log.Println("Env DEFAULT_DIND_IMAGE=" + image)
	log.Println("Env ASSETS_DIR=" + assetsDir)
	log.Println("Env DefaultSessionDuration=" + defaultSessionDuration)

	e := initEvent()
	s := initStorage()
	df := initDockerFactory(s)
	kf := initK8sFactory(s)

	ipf := provisioner.NewInstanceProvisionerFactory(provisioner.NewWindowsASG(df, s), provisioner.NewDinD(id.XIDGenerator{}, df, s))
	sp := provisioner.NewOverlaySessionProvisioner(df)

	core := pwd.NewPWD(df, e, s, sp, ipf)

	tasks := []scheduler.Task{
		task.NewCheckPorts(e, df),
		task.NewCheckSwarmPorts(e, df),
		task.NewCheckSwarmStatus(e, df),
		task.NewCollectStats(e, df, s),
		task.NewCheckK8sClusterStatus(e, kf),
		task.NewCheckK8sClusterExposedPorts(e, kf),
	}
	sch, err := scheduler.NewScheduler(tasks, s, e, core)
	if err != nil {
		log.Fatal("Error initializing the scheduler: ", err)
	}

	sch.Start()

	d, err := time.ParseDuration(config.DefaultSessionDuration)
	if err != nil {
		log.Fatalf("Cannot parse duration %s. Got: %v", config.DefaultSessionDuration, err)
	}

	// todo: update this using env-var
	// update AssetDir to switch between pwd and pwk deployment
	// pwd: default
	// pwk: k8s
	playground := types.Playground{
		Domain: config.PlaygroundDomain,
		DefaultDinDInstanceImage:    config.DefaultDinDImage,
		AllowWindowsInstances:       config.NoWindows,
		DefaultSessionDuration:      d,
		AvailableDinDInstanceImages: []string{config.DefaultDinDImage},
		AssetsDir:                   assetsDir,
		Tasks:                       []string{".*"}}

	if _, err := core.PlaygroundNew(playground); err != nil {
		log.Fatalf("Cannot create default playground. Got: %v", err)
	}

	handlers.Bootstrap(core, e)
	handlers.Register(nil)
}

func initStorage() storage.StorageApi {
	s, err := storage.NewFileStorage(config.SessionsFile)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal("Error initializing StorageAPI: ", err)
	}
	return s
}

func initEvent() event.EventApi {
	return event.NewLocalBroker()
}

func initDockerFactory(s storage.StorageApi) docker.FactoryApi {
	return docker.NewLocalCachedFactory(s)
}

func initK8sFactory(s storage.StorageApi) k8s.FactoryApi {
	return k8s.NewLocalCachedFactory(s)
}
