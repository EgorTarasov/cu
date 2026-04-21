package cli

func init() {
	RootCmd.AddCommand(loginCmd)
	RootCmd.AddCommand(fetchCmd)
	RootCmd.AddCommand(deadlinesCmd)
	RootCmd.AddCommand(gradesCmd)
	RootCmd.AddCommand(materialsCmd)
	RootCmd.AddCommand(taskCmd)
	RootCmd.AddCommand(coursesCmd)
}
