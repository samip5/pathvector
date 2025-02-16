![Banner](https://pathvector.io/img/black-border.svg)

Pathvector is a declarative BGP routing platform that automates route optimization and control plane configuration with secure and repeatable routing policy.

[![Docs](https://img.shields.io/static/v1?label=docs&message=pathvector.io&color=9407cd&style=for-the-badge)](https://pathvector.io)
[![Go Report](https://goreportcard.com/badge/github.com/natesales/pathvector?style=for-the-badge)](https://goreportcard.com/report/github.com/natesales/pathvector)
[![License](https://img.shields.io/github/license/natesales/pathvector?style=for-the-badge)](https://github.com/natesales/pathvector/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/natesales/pathvector?style=for-the-badge)](https://github.com/natesales/pathvector/releases)
[![Codecov](https://img.shields.io/codecov/c/github/natesales/pathvector?style=for-the-badge)](https://app.codecov.io/gh/natesales/pathvector)

## Features

* Robust BGP route filtering with RPKI, IRR, and import limits automatically configured from PeeringDB.
* Automatic route optimization by enriching the standard set of BGP attributes with latency and packet loss metrics.
* Single YAML configuration file: Want to track your changes? Just commit your file to version control.
* Data-plane Agnostic: Pathvector works on servers, network switches, embedded devices, etc.
* Built on Open Source: In addition to Pathvector itself, it's dependencies such as [bird](https://gitlab.nic.cz/labs/bird/), [keepalived](https://github.com/acassen/keepalived), [gortr](https://github.com/cloudflare/gortr) and [bgpq4](https://github.com/bgp/bgpq4) are open source and free to use.
