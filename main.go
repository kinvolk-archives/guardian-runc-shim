package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type Config struct {
	Binary  string `required:"true"`
	LogFile string `default:"/var/log/runc-wrapper"`
}

func modifyConfig(b []byte) ([]byte, error) {
	// Decode JSON.
	var c specs.Spec
	err := json.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}

	// Add required devices for libvirt.
	mode0660 := os.FileMode(0660)
	mode0666 := os.FileMode(0666)
	uid := uint32(0)
	gid := uint32(104)
	d := []specs.LinuxDevice{
		specs.LinuxDevice{
			Path:     "/dev/kvm",
			Type:     "c",
			Major:    10,
			Minor:    232,
			FileMode: &mode0660,
			UID:      &uid,
			GID:      &gid,
		},
		specs.LinuxDevice{
			Path:     "/dev/net/tun",
			Type:     "c",
			Major:    10,
			Minor:    200,
			FileMode: &mode0666,
			UID:      &uid,
			GID:      &gid,
		},
	}
	if c.Linux == nil {
		c.Linux = &specs.Linux{}
	}
	if c.Linux.Devices == nil {
		c.Linux.Devices = []specs.LinuxDevice{}
	}
	c.Linux.Devices = append(c.Linux.Devices, d...)

	// Make /sys mount read/write.
	for _, m := range c.Mounts {
		if m.Destination == "/sys" {
			for i, o := range m.Options {
				if o == "ro" {
					m.Options[i] = "rw"
				}
			}
		}
	}

	// Whitelist all devices.
	if c.Linux.Resources == nil {
		c.Linux.Resources = &specs.LinuxResources{}
	}
	if c.Linux.Resources.Devices == nil {
		c.Linux.Resources.Devices = []specs.LinuxDeviceCgroup{}
	}
	c.Linux.Resources.Devices = []specs.LinuxDeviceCgroup{
		specs.LinuxDeviceCgroup{Allow: true, Access: "rwm"},
	}

	// Encode JSON.
	res, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func main() {
	var c Config
	err := envconfig.Process("runc_wrapper", &c)
	if err != nil {
		log.Fatal(err)
	}

	// Set up logging.
	f, err := os.OpenFile(c.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// Verify runc binary exists.
	binPath, err := exec.LookPath(c.Binary)
	if err != nil {
		log.Fatalf("Binary '%s' does not exist\n", c.Binary)
	}
	log.Printf("Using runc binary at %s\n", binPath)

	sArgs := strings.Join(os.Args[1:], " ")
	log.Printf("Wrapper called with arguments: %s\n", sArgs)

	// When true we need to modify config.json.
	var modify bool

	// Check if `runc run` was called.
	for _, a := range os.Args[1:] {
		if a == "run" {
			modify = true
		}
	}

	if modify {
		// Extract container ID and bundle path.
		var cid, bundle string
		for i, v := range os.Args[1:] {
			if v == "--bundle" {
				// Take the element right after "--bundle". i+2 because we are iterating over
				// os.Args[1:], not os.Args.
				bundle = os.Args[i+2]
			}
		}
		cid = os.Args[len(os.Args)-1]

		if bundle == "" {
			log.Fatal("Could not read bundle path")
		}

		if cid == "" {
			log.Fatal("Could not read container ID")
		}

		log.Printf("Bundle path: %s\n", bundle)
		log.Printf("Container ID: %s\n", cid)

		log.Println("Modifying config.json")
		cf := filepath.Join(bundle, "config.json")
		config, err := ioutil.ReadFile(cf)
		if err != nil {
			log.Fatalf("Reading config file: %v\n", err)
		}

		newConfig, err := modifyConfig(config)
		if err != nil {
			log.Fatalf("Modifying config: %v\n", err)
		}

		// Overwrite config.
		log.Println("Writing new config")
		err = ioutil.WriteFile(cf, newConfig, 0600)
		if err != nil {
			log.Fatalf("Writing new config: %v\n", err)
		}
	}

	log.Printf("Executing %s %s\n", binPath, sArgs)
	err = syscall.Exec(binPath, append([]string{binPath}, os.Args[1:]...), os.Environ())
	if err != nil {
		log.Fatal(err)
	}
}
