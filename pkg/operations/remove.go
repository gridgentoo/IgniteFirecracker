package operations

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	api "github.com/weaveworks/ignite/pkg/apis/ignite"
	meta "github.com/weaveworks/ignite/pkg/apis/meta/v1alpha1"
	"github.com/weaveworks/ignite/pkg/client"
	"github.com/weaveworks/ignite/pkg/logs"
	"github.com/weaveworks/ignite/pkg/providers"
	"github.com/weaveworks/ignite/pkg/runtime"
	"github.com/weaveworks/ignite/pkg/util"
)

const (
	signalSIGQUIT = "SIGQUIT"
)

// DeleteVM removes the specified VM from the Client and performs a cleanup
func DeleteVM(c *client.Client, vm *api.VM) error {
	if err := c.VMs().Delete(vm.GetUID()); err != nil {
		return err
	}

	return CleanupVM(vm)
}

// CleanupVM removes the resources of the given VM
func CleanupVM(vm *api.VM) error {
	// Inspect the container before trying to stop it and it gets auto-removed
	result, _ := providers.Runtime.InspectContainer(util.NewPrefixer().Prefix(vm.GetUID()))

	// If the VM is running, try to kill it first so we don't leave dangling containers
	if vm.Running() {
		if err := StopVM(vm, true, true); err != nil {
			return err
		}
	}

	// Remove the VM container if it exists
	if result != nil {
		if err := RemoveVMContainer(vm, result); err != nil {
			return err
		}
	}

	if logs.Quiet {
		fmt.Println(vm.GetUID())
	} else {
		log.Infof("Removed %s with name %q and ID %q", vm.GetKind(), vm.GetName(), vm.GetUID())
	}

	return nil
}

func RemoveVMContainer(vm meta.Object, result *runtime.ContainerInspectResult) error {
	// Remove the VM container. If the container has been/is being automatically removed
	// between InspectContainer and this call, RemoveContainer will throw an error. Currently
	// we have no real way to inspect and remove immediately without having a potential race
	// condition, so ignore the error for now. TODO: Robust conditional remove support
	_ = providers.Runtime.RemoveContainer(result.ID)

	// Tear down the networking of the VM
	return removeNetworking(vm.(*api.VM), result.ID)
}

// StopVM stops or kills a VM
func StopVM(vm *api.VM, kill, silent bool) error {
	var err error
	container := util.NewPrefixer().Prefix(vm.GetUID())
	action := "stop"

	// Stop or kill the VM container
	if kill {
		action = "kill"
		err = providers.Runtime.KillContainer(container, signalSIGQUIT) // TODO: common constant for SIGQUIT
	} else {
		err = providers.Runtime.StopContainer(container, nil)
	}

	if err != nil {
		return fmt.Errorf("failed to %s container for %s %q: %v", action, vm.GetKind(), vm.GetUID(), err)
	}

	if silent {
		return nil
	}

	if logs.Quiet {
		fmt.Println(vm.GetUID())
	} else {
		log.Infof("Stopped %s with name %q and ID %q", vm.GetKind(), vm.GetName(), vm.GetUID())
	}

	return nil
}

func removeNetworking(vm *api.VM, containerID string) error {
	// Perform the removal
	networkPlugin := providers.NetworkPlugins[vm.Spec.Network.Mode.String()]
	log.Debugf("Removing the container with ID %q from the %q network", containerID, networkPlugin.Name())
	return networkPlugin.RemoveContainerNetwork(containerID)
}
