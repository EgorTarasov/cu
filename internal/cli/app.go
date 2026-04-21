package cli

func init() {
	RootCmd.AddCommand(coursesCmd)
	RootCmd.AddCommand(fetchCmd)
	RootCmd.AddCommand(loginCmd)
}
