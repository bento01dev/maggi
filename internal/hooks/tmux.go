package hooks

import "fmt"

func AddHook(hookType string, globalHook bool) error {
	fmt.Printf("adding hook, hook type: %s, global: %v\n", hookType, globalHook)
    //TODO: need to attach hooks for session-created, window-linked, pane-focus-in
    // focus-events needs to be turned on for pane-focus-in
	return nil
}

func DeleteHook(hookType string, globalHook bool) error {
	fmt.Printf("delete hook, hook type: %s, global: %v\n", hookType, globalHook)
	return nil
}
