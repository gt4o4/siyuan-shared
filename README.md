<p align="center">
<img alt="SiYuan" src="https://b3log.org/images/brand/siyuan-128.png">
<br>
<em>Refactor your thinking</em>
<br><br>
<a title="Releases" target="_blank" href="https://github.com/gt4o4/siyuan-shared/releases"><img src="https://img.shields.io/github/release/gt4o4/siyuan-shared.svg?style=flat-square&color=9CF"></a>
<a title="Downloads" target="_blank" href="https://github.com/gt4o4/siyuan-shared/releases"><img src="https://img.shields.io/github/downloads/gt4o4/siyuan-shared/total.svg?style=flat-square&color=blueviolet"></a>
<a title="Docker Pulls" target="_blank" href="https://hub.docker.com/r/apkdv/siyuan-shared"><img src="https://img.shields.io/docker/pulls/apkdv/siyuan-shared.svg?style=flat-square&color=green"></a>
<a title="Docker Image Size" target="_blank" href="https://hub.docker.com/r/apkdv/siyuan-shared"><img src="https://img.shields.io/docker/image-size/apkdv/siyuan-shared.svg?style=flat-square&color=ff96b4"></a>
<br>
<a title="AGPLv3" target="_blank" href="https://www.gnu.org/licenses/agpl-3.0.txt"><img src="http://img.shields.io/badge/license-AGPLv3-orange.svg?style=flat-square"></a>
<a title="Discord" target="_blank" href="https://discord.gg/dmMbCqVX7G"><img alt="Chat on Discord" src="https://img.shields.io/discord/808152298789666826?label=Discord&logo=Discord&style=social"></a>
</p>

<p align="center">
<a href="README_zh_CN.md">中文</a>
</p>

---

## Table of Contents

* [💡 Introduction](#-introduction)
* [🔮 Features](#-features)
* [🏗️ Architecture and Ecosystem](#-architecture-and-ecosystem)
* [🚀 Download](#-download)
* [🐳 Docker Hosting](#-docker-hosting)

---

## 💡 Introduction

A community fork of [SiYuan Note](https://github.com/siyuan-note/siyuan) — a privacy-first personal knowledge management system with fine-grained block-level reference and Markdown WYSIWYG.

This fork runs the Go kernel as an in-process N-API addon (c-shared library linked into Electron via Node.js native module), eliminating the separate kernel process. It also provides S3/WebDAV sync, self-hosted Docker deployment with all features enabled, and tracks upstream SiYuan releases.

![feature0.png](https://b3logfile.com/file/2025/11/feature0-GfbhEqf.png)

![feature51.png](https://b3logfile.com/file/2025/11/feature5-1-7DJSfEP.png)

## 🔮 Features

Most features are free, even for commercial use.

* Content block
  * Block-level reference and two-way links
  * Custom attributes
  * SQL query embed
  * Protocol `siyuan://`
* Editor
  * Block-style
  * Markdown WYSIWYG
  * List outline
  * Block zoom-in
  * Million-word large document editing
  * Mathematical formulas, charts, flowcharts, Gantt charts, timing charts, staffs, etc.
  * Web clipping
  * PDF Annotation link
* Export
  * Block ref and embed
  * Standard Markdown with assets
  * PDF, Word and HTML
  * Copy to WeChat MP, Zhihu and Yuque
* Database
  * Table view
* Flashcard spaced repetition
* AI writing and Q/A chat via OpenAI API
* Tesseract OCR
* Multi-tab, drag and drop to split screen
* Template snippet
* JavaScript/CSS snippet
* Android/iOS/HarmonyOS App
* Docker deployment
* [API](API.md)
* Community marketplace

## 🏗️ Architecture and Ecosystem

![SiYuan Arch](https://b3logfile.com/file/2023/05/SiYuan_Arch-Sgu8vXT.png "SiYuan Arch")

| Project                                                  | Description           | Forks                                                                           | Stars                                                                                |
|----------------------------------------------------------|-----------------------|---------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| [lute](https://github.com/88250/lute)                    | Editor engine         | ![GitHub forks](https://img.shields.io/github/forks/88250/lute)                 | ![GitHub Repo stars](https://img.shields.io/github/stars/88250/lute)                 |
| [chrome](https://github.com/siyuan-note/siyuan-chrome)   | Chrome/Edge extension | ![GitHub forks](https://img.shields.io/github/forks/siyuan-note/siyuan-chrome)  | ![GitHub Repo stars](https://img.shields.io/github/stars/siyuan-note/siyuan-chrome)  |
| [bazaar](https://github.com/siyuan-note/bazaar)          | Community marketplace | ![GitHub forks](https://img.shields.io/github/forks/siyuan-note/bazaar)         | ![GitHub Repo stars](https://img.shields.io/github/stars/siyuan-note/bazaar)         |
| [dejavu](https://github.com/siyuan-note/dejavu)          | Data repo             | ![GitHub forks](https://img.shields.io/github/forks/siyuan-note/dejavu)         | ![GitHub Repo stars](https://img.shields.io/github/stars/siyuan-note/dejavu)         |
| [petal](https://github.com/siyuan-note/petal)            | Plugin API            | ![GitHub forks](https://img.shields.io/github/forks/siyuan-note/petal)          | ![GitHub Repo stars](https://img.shields.io/github/stars/siyuan-note/petal)          |
| [android](https://github.com/siyuan-note/siyuan-android) | Android App           | ![GitHub forks](https://img.shields.io/github/forks/siyuan-note/siyuan-android) | ![GitHub Repo stars](https://img.shields.io/github/stars/siyuan-note/siyuan-android) |
| [ios](https://github.com/siyuan-note/siyuan-ios)         | iOS App               | ![GitHub forks](https://img.shields.io/github/forks/siyuan-note/siyuan-ios)     | ![GitHub Repo stars](https://img.shields.io/github/stars/siyuan-note/siyuan-ios)     |
| [harmony](https://github.com/siyuan-note/siyuan-harmony) | HarmonyOS App         | ![GitHub forks](https://img.shields.io/github/forks/siyuan-note/siyuan-harmony) | ![GitHub Repo stars](https://img.shields.io/github/stars/siyuan-note/siyuan-harmony) |
| [riff](https://github.com/siyuan-note/riff)              | Spaced repetition     | ![GitHub forks](https://img.shields.io/github/forks/siyuan-note/riff)           | ![GitHub Repo stars](https://img.shields.io/github/stars/siyuan-note/riff)           |

## 🚀 Download

[GitHub Releases](https://github.com/gt4o4/siyuan-shared/releases)

Available builds: Linux (AppImage, tar.gz), macOS (Intel & ARM64 DMG), Windows (exe), Android (APK), iOS (IPA).

## 🐳 Docker Hosting

<details>
<summary>Docker Deployment</summary>

#### Overview

The easiest way to serve SiYuan on a server is to deploy it through Docker.

* Image name `apkdv/siyuan-shared`
* [Image URL](https://hub.docker.com/r/apkdv/siyuan-shared)

#### File structure

The overall program is located under `/opt/siyuan/`, which is basically the structure under the resources folder of the Electron installation package:

* appearance: icon, theme, languages
* guide: user guide document
* stage: interface and static resources
* kernel: kernel program

#### Entrypoint

The entry point is set when building the Docker image: `ENTRYPOINT ["/opt/siyuan/entrypoint.sh"]`. This script allows changing the `PUID` and `PGID` of the user that will run inside the container. This is especially relevant to solve permission issues when mounting directories from the host.

Use the following parameters when running the container with `docker run apkdv/siyuan-shared`:

* `--workspace`: Specifies the workspace folder path, mounted to the container via `-v` on the host
* `--accessAuthCode`: Specifies the access authorization code

More parameters can be found using `--help`. Here's an example startup command:

```bash
docker run -d \
  -v workspace_dir_host:workspace_dir_container \
  -p 6806:6806 \
  -e PUID=1001 -e PGID=1002 \
  apkdv/siyuan-shared \
  --workspace=workspace_dir_container \
  --accessAuthCode=xxx
```

* `PUID`: Custom user ID (optional, defaults to `1000` if not provided)
* `PGID`: Custom group ID (optional, defaults to `1000` if not provided)
* `workspace_dir_host`: The workspace folder path on the host
* `workspace_dir_container`: The path of the workspace folder in the container, as specified in `--workspace`
  * Alternatively, set the path via the `SIYUAN_WORKSPACE_PATH` env variable. The command line value takes priority if both are set
* `accessAuthCode`: Access authorization code (please **be sure to modify**, otherwise anyone can access your data)
  * Alternatively, set the auth code via the `SIYUAN_ACCESS_AUTH_CODE` env variable. The command line value takes priority if both are set
  * To disable the access authorization code, set the env variable `SIYUAN_ACCESS_AUTH_CODE_BYPASS=true`

To simplify things, it is recommended to configure the workspace folder path to be consistent on the host and container, such as having both set to `/siyuan/workspace`:

```bash
docker run -d \
  -v /siyuan/workspace:/siyuan/workspace \
  -p 6806:6806 \
  -e PUID=1001 -e PGID=1002 \
  apkdv/siyuan-shared \
  --workspace=/siyuan/workspace/ \
  --accessAuthCode=xxx
```

#### Docker Compose

```yaml
version: "3.9"
services:
  main:
    image: apkdv/siyuan-shared
    command: ['--workspace=/siyuan/workspace/', '--accessAuthCode=${AuthCode}']
    ports:
      - 6806:6806
    volumes:
      - /siyuan/workspace:/siyuan/workspace
    restart: unless-stopped
    environment:
      # A list of time zone identifiers can be found at https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
      - TZ=${YOUR_TIME_ZONE}
      - PUID=${YOUR_USER_PUID}
      - PGID=${YOUR_USER_PGID}
```

#### User Permissions

The `entrypoint.sh` script ensures the creation of the `siyuan` user and group with the specified `PUID` and `PGID`. When the host creates a workspace folder, set the user and group ownership to match. For example:

```bash
chown -R 1001:1002 /siyuan/workspace
```

#### Hidden port

Use NGINX reverse proxy to hide port 6806. Note:

* Configure WebSocket reverse proxy for `/ws`

#### Note

* Be sure to confirm the correctness of the mounted volume, otherwise data will be lost after the container is deleted
* Do not use URL rewriting for redirection, otherwise there may be problems with authentication. Configure a reverse proxy instead

#### Limitations

* Does not support desktop and mobile application connections, only supports use on browsers
* Export to PDF, HTML and Word formats is not supported
* Import Markdown file is not supported

</details>
