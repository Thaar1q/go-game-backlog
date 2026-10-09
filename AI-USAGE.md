# AI Usage Disclosure

This document outlines the scope, tools, and practices regarding the use of Artificial Intelligence (AI) assistance in the development of this project.

## 1. Tools Used
- **Large Language Models (LLM):** Gemini (Google) / GitHub Copilot
- **Development Environment:** Visual Studio Code, Windows PowerShell, Git Bash
- **Purpose:** Syntax reference, debugging edge cases, CLI troubleshooting, and document formatting.

## 2. Scope of AI Assistance
- **Syntax & Validation Reference:** Struct tag patterns for `validator/v10` (`omitnil` on PATCH pointer fields).
- **CLI & Script Debugging:** Troubleshooting PowerShell string escaping and curl commands for automated test runners (`data_test_uts.ps1`).
- **Edge Case Consultation:** Validating expected HTTP error codes for edge scenarios (e.g., self-demotion 422, replay token 401).
- **Documentation:** Assisting in Markdown table formatting and technical report structure.

## 3. Human Oversight & Verification
- **Code Ownership:** All application architecture, database queries, domain logic, and routes were authored, understood, and maintained by the developer.
- **Testing:** All endpoints and unit tests were executed and validated locally by the developer against course specifications.
