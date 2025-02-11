package main

var RBAC = make(map[string][]int)

func RBACinit() {
	RBAC["Start"] = []int{1, 2}
	RBAC["Stop"] = []int{1, 2}
	RBAC["List"] = []int{1, 2}
	RBAC["Status"] = []int{1, 2}
	RBAC["Running"] = []int{1}
	RBAC["Add"] = []int{1}
	RBAC["Auth"] = []int{0, 1, 2}

}

func ChackRole(slice []int, num int) bool {
	for _, v := range slice {
		if v == num {
			return true
		}
	}
	return false
}

func CheckPermission(m map[string][]int, num int) []string {
	var keys []string
	// Going through all the keys and values of the map
	for key, values := range m {

		// Check if there is a number in the array of values
		for _, v := range values {
			if v == num {
				keys = append(keys, key) // Add the key to the result
				break                    // Interrupt the inner loop since we found the desired number
			}
		}
	}
	return keys
}
