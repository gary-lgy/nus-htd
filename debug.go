package main

import "fmt"

const emptyPlaceholer string = "<empty>"

func debugPrint(command string, username, password, morningOrAfternoon *string, temperature *float32, hasSymptoms, reportAnomaly *bool) {
	fmt.Printf("Command: %s\n", command)
	fmt.Printf("Username: %s\n", getString(username))
	fmt.Printf("Password: %s\n", getString(password))
	fmt.Printf("morningOrAfternoon: %s\n", getString(morningOrAfternoon))
	fmt.Printf("temperature: %s\n", getFloat(temperature))
	fmt.Printf("hasSymptoms: %s\n", getBool(hasSymptoms))
	fmt.Printf("reportAnomaly: %s\n", getBool(reportAnomaly))
}

func getString(ptr *string) string {
	if ptr == nil {
		return emptyPlaceholer
	}
	return *ptr
}

func getFloat(ptr *float32) string {
	if ptr == nil {
		return emptyPlaceholer
	}
	return fmt.Sprintf("%.1f", *ptr)
}

func getBool(ptr *bool) string {
	if ptr == nil {
		return emptyPlaceholer
	}
	return fmt.Sprintf("%v", *ptr)
}
