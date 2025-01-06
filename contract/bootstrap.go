package contract

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const cargoTemplate = `[package]
name = "%s"
version = "0.1.0"
edition = "2021"

[build]
target = "wasm32-unknown-unknown"

[lib]
crate-type = ["cdylib", "rlib"]

[dependencies]
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
rubixwasm-std = { git = "https://github.com/rubixchain/rubix-wasm.git", subdir = "packages/std" }
`

const libTemplate = `use rubixwasm_std::{errors::WasmError, contract_fn};
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
pub struct AddThreeNumsReq {
    pub a: u32,
    pub b: u32,
    pub c: u32,
}

// A sample smart contract function that adds three numbers.
// Every Rubix Smart Contract function expects a single struct input,
// and the output must be of type Result<String, WasmError>
#[contract_fn]
pub fn add_three_nums(input: AddThreeNumsReq) -> Result<String, WasmError> {
    if input.b == 0 {
        return Err(WasmError::from("Parameter 'b' cannot be zero"))
    }

    let sum = input.a + input.b + input.c;
    Ok(sum.to_string())
}
`

// Bootstrap creates a new Rust smart contract project with the given name
func Bootstrap(name string) error {
	// Validate contract name
	if name == "" {
		return fmt.Errorf("contract name cannot be empty")
	}
	if !isValidRustName(name) {
		return fmt.Errorf("invalid contract name: must be a valid Rust package name (lowercase alphanumeric with hyphens)")
	}

	// Create project directory
	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create src directory
	srcDir := filepath.Join(name, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return fmt.Errorf("failed to create src directory: %w", err)
	}

	// Create Cargo.toml
	cargoContent := fmt.Sprintf(cargoTemplate, name)
	if err := os.WriteFile(filepath.Join(name, "Cargo.toml"), []byte(cargoContent), 0644); err != nil {
		return fmt.Errorf("failed to create Cargo.toml: %w", err)
	}

	// Create src/lib.rs
	if err := os.WriteFile(filepath.Join(srcDir, "lib.rs"), []byte(libTemplate), 0644); err != nil {
		return fmt.Errorf("failed to create lib.rs: %w", err)
	}

	return nil
}

// isValidRustName checks if the given name is a valid Rust package name
func isValidRustName(name string) bool {
	if name == "" {
		return false
	}

	// Must be lowercase
	if name != strings.ToLower(name) {
		return false
	}

	// Can only contain alphanumeric and hyphens
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}

	// Cannot start or end with hyphen
	if name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}

	return true
}
