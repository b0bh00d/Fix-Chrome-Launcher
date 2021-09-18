package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kardianos/service"
	"golang.org/x/sys/windows/registry"
)

const (
    default_polling_interval = 60 // check every N seconds
)

var logger service.Logger

type program struct {
    exit chan struct{}
}

func (p *program) Start(s service.Service) error {
    p.exit = make(chan struct{})

    go p.run()

    return nil
}

func (p *program) run() error {
    var critical_options map[string]interface{}

    polling_interval := default_polling_interval

    key, err := registry.OpenKey(registry.CLASSES_ROOT, `ChromeHTML\shell\open\command`, registry.QUERY_VALUE)
    if err != nil {
        logger.Error(err)
    } else {
        // first, see if they want to override the default polling interval
        if interval, _, err := key.GetIntegerValue("fcl_interval"); err == nil {
            polling_interval = int(interval)
        }

        // next, get the "sticky" options for the launcher string
        if j1, _, err := key.GetStringValue("fcl_options"); err != nil {
            logger.Error(err)
        } else
        {
            json_bytes := []byte(j1);
            if err := json.Unmarshal(json_bytes, &critical_options); err != nil {
                logger.Error(err)
            }
        }
    }

    key.Close();

    if len(critical_options) == 0 {
        logger.Warning("Sticky options could not be read from the registry; nothing to do!")
        return nil
    }

    logger.Infof("Running with a polling interval of %d second.", polling_interval)

    ticker := time.NewTicker(time.Duration(polling_interval) * time.Second)

    for {
        select {
        case <-ticker.C:
            key, err := registry.OpenKey(registry.CLASSES_ROOT, `ChromeHTML\shell\open\command`, registry.QUERY_VALUE|registry.SET_VALUE)

            if err != nil {
                logger.Error(err)
            } else {
                if launch_str, _, err := key.GetStringValue(""); err != nil {
                    logger.Error(err)
                } else {
                    if ndx := strings.Index(launch_str, "--single"); ndx == -1 {
                        logger.Error("Could not locate '--single-argument' in Chrome launcher string!")
                    } else {
                        do_update := false
                        for key, val := range critical_options {
                            if arg, ok := val.(string); !ok {
                                logger.Warning("Could not convert 'arg' value to string!")
                            } else
                            {
                                if !strings.Contains(launch_str, key) {
                                    ndx = strings.Index(launch_str, "--single")
                                    s := fmt.Sprintf("%s%s %s", launch_str[:ndx], key, launch_str[ndx:])
                                    launch_str = s

                                    if len(arg) != 0 {
                                        ndx = strings.Index(launch_str, "--single")
                                        s := fmt.Sprintf("%s%s %s", launch_str[:ndx], arg, launch_str[ndx:])
                                        launch_str = s
                                    }

                                    do_update = true
                                }
                            }
                        }

                        if do_update {
                            logger.Infof("Setting Chrome launcher string: %s...", launch_str)
                            if err = key.SetStringValue("", launch_str); err != nil {
                                logger.Error(err)
                            }
                        }
                    }
                }
            }

            key.Close();

        case <-p.exit:
            ticker.Stop()
            return nil
        }
    }
}

func (p *program) Stop(s service.Service) error {
    close(p.exit)
    return nil
}

func main() {
    svcFlag := flag.String("service", "", "Control the system service.")
    flag.Parse()

    options := make(service.KeyValue)
    options["Restart"] = "on-success"
    options["SuccessExitStatus"] = "1 2 8 SIGKILL"
    svcConfig := &service.Config{
        Name:         "FixChromeLauncher",
        DisplayName:  "Fix Chrome Launcher",
        Description:  "A service to keep Chrome from destroying custom launch options.",
        Dependencies: []string{},
        Option:       options,
    }

    prg := &program{}
    s, err := service.New(prg, svcConfig)
    if err != nil {
        log.Fatal(err)
    }
    errs := make(chan error, 5)
    logger, err = s.Logger(errs)
    if err != nil {
        log.Fatal(err)
    }

    go func() {
        for {
            err := <-errs
            if err != nil {
                log.Print(err)
            }
        }
    }()

    if len(*svcFlag) != 0 {
        err := service.Control(s, *svcFlag)
        if err != nil {
            log.Printf("Valid actions: %q\n", service.ControlAction)
            log.Fatal(err)
        }
        return
    }
    err = s.Run()
    if err != nil {
        logger.Error(err)
    }
}
