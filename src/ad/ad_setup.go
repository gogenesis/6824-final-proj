package ad

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Envionment variable names ===================================================

var envVarNamePrefix = "DFS_" // so that all env variables start with the same thing. Set to "" to disable.
var disableAssertionsEnvVarName = envVarNamePrefix + "DISABLE_ASSERTIONS"
var defaultDebugLevelEnvVarName = envVarNamePrefix + "DEFAULT_DEBUG_LEVEL"

// e.g, the debug level for the fsraft package is named DFS_FSRAFT_DEBUG_LEVEL
func packageDebugLevelEnvVarName(pkg string) string {
	return envVarNamePrefix + strings.ToUpper(pkg) + "_DEBUG_LEVEL"
}

// Package-wide variables =====================================================

var packageNamesToDebugLevels = map[string]int{
	// -1 = unset
	"raft":     -1,
	"fsraft":   -1,
	"memoryfs": -1,
}

var assertionsEnabled bool

// if no environment variables are set
var defaultDebugLevel = NONE

// The actual setup function ===================================================

func init() {
	// Will be automatically run at the beginning of every run.

	assertionsEnabled = true
	disableAssertionsVar := os.Getenv(disableAssertionsEnvVarName)
	if strings.ToLower(disableAssertionsVar) == "true" {
		fmt.Printf("Disabling assertions because $%v==true\n", disableAssertionsEnvVarName)
		assertionsEnabled = false
	}

	for packageName := range packageNamesToDebugLevels {
		envVarName := packageDebugLevelEnvVarName(packageName)
		intValue, isInt := envVarIntValueAndIsInt(envVarName)
		if !isInt {
			continue
		}
		if intValue < 0 || intValue >= len(loggingLevelNames) {
			fmt.Printf("Environment variable %v tried to set debug level of package %v to %d, but "+
				"valid debug levels are 0 through %d inclusive. Ignoring it.\n", envVarName, packageName,
				intValue, len(loggingLevelNames)-1)
			continue
		}
		packageNamesToDebugLevels[packageName] = intValue
		fmt.Printf("Setting debug level for package %v to %v through environment variable %v.\n",
			packageName, debugLevelName(packageNamesToDebugLevels[packageName]), envVarName)
	}

	debugLevelForOtherPackages, setWithEnvVar := envVarIntValueAndIsInt(defaultDebugLevelEnvVarName)
	var explanation = " through default environment variable " + defaultDebugLevelEnvVarName
	if !setWithEnvVar {
		debugLevelForOtherPackages = defaultDebugLevel
		explanation = " because it was not set through any environment variable" //
	}
	for packageName := range packageNamesToDebugLevels {
		// If it's not set above
		if packageNamesToDebugLevels[packageName] == -1 {
			packageNamesToDebugLevels[packageName] = debugLevelForOtherPackages
			fmt.Printf("Setting debug level for package %-8v to %v%v.\n",
				packageName, debugLevelName(packageNamesToDebugLevels[packageName]), explanation)
		}
	}

	// Make sure we didn't forget anything
	for _, val := range packageNamesToDebugLevels {
		Assert(val != -1)
	}
}

func envVarIntValueAndIsInt(envVarName string) (intValue int, isInt bool) {
	envVarValue, envVarIsSet := os.LookupEnv(envVarName)
	if !envVarIsSet {
		return 0, false // exit peacefully without complaint
	}
	int64Value, err := strconv.ParseInt(envVarValue, 10, 8) // base 10, 8-bit integer
	if err != nil {
		fmt.Printf("Tried to parse environment variable %v as a number. Ignoring it.\n",
			envVarName)
		return 0, false
	}
	return int(int64Value), true
}
