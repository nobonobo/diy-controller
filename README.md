# DIY Steering Wheel Controller Firmware

![License](https://img.shields.io/badge/license-MIT-blue.svg)

## Overview
This repository contains the firmware for a custom DIY steering wheel controller. The firmware runs on a microcontroller, provides low‑latency, high‑precision input handling for racing simulators, and supports force‑feedback.

## Features
- **USB HID** support for plug‑and‑play operation on Windows, Linux, and macOS.

## Supported Devices
- **Required**: [RP2040+CAN](https://ssci.to/9279)
- **Optional (choose one)**: [For Steering Style Motor (High Torque)](https://ssci.to/9219) or [For Handheld Dial Style Motor](https://ssci.to/10027)

## Hardware Requirements
- A compatible microcontroller board (e.g., RP2040 or similar).
- USB Type‑C or Micro‑USB connector for wired operation.

## Prerequisites
- Docker (container runtime)
- `task` command (Taskfile support)
- Go (development)
- TinyGo (firmware compilation)
- No additional tools required (e.g., Python, CMake, OpenOCD are unnecessary)

## Build Instructions
```bash
task docker
```
The build will produce UF2 files in the `dist/` directory.

## Flashing the Firmware
1. Put the target device into write (bootloader) mode. It will appear as a USB mass storage device named **RPI-RP2**.
2. On Windows, simply drag and drop the generated `.uf2` file from the `dist/` folder onto the mounted drive.

## Contributing
Contributions are welcome! Please follow these steps:
1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/your-feature`).
3. Ensure code style compliance with `clang-format` and run `make lint`.
4. Submit a Pull Request with a clear description of the changes.

## License
This project is licensed under the MIT License – see the [LICENSE](LICENSE) file for details.

---
*Crafted with passion for the racing community.*
