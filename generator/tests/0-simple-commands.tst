Test: Simple commands
------------

--------input--------
# Nothing but this comment
# we need to make sure empty shell scripts produce empty go programs


--------output--------
package main

import "bunster-build/runtime"

func Main(shell *runtime.Shell) {
}

------------

--------input--------
git
--------output--------
package main

import "bunster-build/runtime"

func Main(shell *runtime.Shell) {
    func() {
  		var cmd_0_name = `git`
  		var cmd_0_args []string
  		var cmd_0 = shell.Command(cmd_0_name, cmd_0_args...)
  		cmd_0_fdt, err := shell.CloneFDT()
  		if err != nil {
 			shell.HandleError("", err)
 			return
  		}
  		defer cmd_0_fdt.Destroy()
  		cmd_0.Stdin = cmd_0_fdt.Get(`0`)
  		cmd_0.Stdout = cmd_0_fdt.Get(`1`)
  		cmd_0.Stderr = cmd_0_fdt.Get(`2`)
  		if err := cmd_0.Run(); err != nil {
 			shell.HandleError(cmd_0_name, err)
 			return
  		}
  		shell.ExitCode = cmd_0.ProcessState.ExitCode()

   	}()
}


------------

--------input--------
$varname
--------output--------
package main

import "bunster-build/runtime"

func Main(shell *runtime.Shell) {
    func() {
  		var cmd_0_name = shell.ReadVar("varname")
  		var cmd_0_args []string
  		var cmd_0 = shell.Command(cmd_0_name, cmd_0_args...)
  		cmd_0_fdt, err := shell.CloneFDT()
  		if err != nil {
 			shell.HandleError("", err)
 			return
  		}
  		defer cmd_0_fdt.Destroy()
  		cmd_0.Stdin = cmd_0_fdt.Get(`0`)
  		cmd_0.Stdout = cmd_0_fdt.Get(`1`)
  		cmd_0.Stderr = cmd_0_fdt.Get(`2`)
  		if err := cmd_0.Run(); err != nil {
 			shell.HandleError(cmd_0_name, err)
 			return
  		}
  		shell.ExitCode = cmd_0.ProcessState.ExitCode()

   	}()
}

--------input--------
"quoted-command"
--------output--------
package main

import "bunster-build/runtime"

func Main(shell *runtime.Shell) {
    func() {
  		var cmd_0_name = `quoted-command`
  		var cmd_0_args []string
  		var cmd_0 = shell.Command(cmd_0_name, cmd_0_args...)
  		cmd_0_fdt, err := shell.CloneFDT()
  		if err != nil {
 			shell.HandleError("", err)
 			return
  		}
  		defer cmd_0_fdt.Destroy()
  		cmd_0.Stdin = cmd_0_fdt.Get(`0`)
  		cmd_0.Stdout = cmd_0_fdt.Get(`1`)
  		cmd_0.Stderr = cmd_0_fdt.Get(`2`)
  		if err := cmd_0.Run(); err != nil {
 			shell.HandleError(cmd_0_name, err)
 			return
  		}
  		shell.ExitCode = cmd_0.ProcessState.ExitCode()

   	}()
}
