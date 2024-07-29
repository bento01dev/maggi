package search

import "fmt"

func Run(profileStr, queryStr string) error {
	fmt.Printf("searching for %s in %s", queryStr, profileStr)
	return nil
}
