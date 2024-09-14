Maggi - Tool for managing env and aliases with tmux 

Use `maggi ui` to manage profiles.

Set values for a profile using `eval $(maggi generate --profile <profile_name>)`

Using with tmux, set env and alias based on session name with a default profile (besides the session name) like `eval $(maggi apply-session --default <default_profile>)`.
This will pick up the session name as another profile to apply. if there is no active tmux session, it will set only the default profile.
The values in profile matching tmux session name will override the values from the default profile.
So this can be set in .zprofile to pick up defaults for normal shell and have session based overrides for tmux session shell.
