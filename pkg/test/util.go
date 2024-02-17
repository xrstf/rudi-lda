package test

import "os"

func TempScript(script string) (string, error) {
	f, err := os.CreateTemp(os.TempDir(), "rudilda")
	if err != nil {
		return "", err
	}

	_, err = f.WriteString(script)
	f.Close()

	if err != nil {
		os.Remove(f.Name())
		return "", err
	}

	return f.Name(), nil
}
