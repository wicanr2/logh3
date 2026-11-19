# logh3 Android 工具鏈 20261119-r1（規格161）。
# 第一個 RUN 與 ~/cht/rich2/tools/docker/android.Dockerfile 逐字相同，讓 Docker 重用已有的
# SDK／NDK／Gradle 層，不必重新下載（目錄名 /opt/rich2-toolchain 因此沿用，只存版本清單）。
# ebitenmobile 必須與 go.mod 的 Ebitengine 同版（v2.8.8），不能沿用 rich2 的 v2.9.9。
# Go 換成 1.24.13（與桌面建置的 golang:1.24-bookworm 同一份）：ebitenmobile v2.8.8 依賴的
# golang.org/x/tools v0.25.0 在 Go 1.26 編不過（tokeninternal: invalid array length）。
# 僅含公開建置工具；原版資料與憑證不在建置 context 或映像內。
FROM golang:1.26.7-bookworm@sha256:659cc38c1a394eeb4dd7e31fff6df128bd33444dcc7afd70e3bed5225749dbc0

ARG ANDROID_CMDLINE_TOOLS_URL=https://dl.google.com/android/repository/commandlinetools-linux-15859902_latest.zip
ARG ANDROID_CMDLINE_TOOLS_SHA256=4e4c464f145a7512b57d088ac6c278c03c9eea610886b35a5e0804e74eedf583
ARG GRADLE_URL=https://services.gradle.org/distributions/gradle-8.9-bin.zip
ARG GRADLE_SHA256=d725d707bfabd4dfdc958c624003b3c80accc03f7037b5122c4b1d0ef15cecab

ENV ANDROID_HOME=/opt/android-sdk \
    ANDROID_SDK_ROOT=/opt/android-sdk \
    ANDROID_NDK_HOME=/opt/android-sdk/ndk/27.2.12479018 \
    GRADLE_HOME=/opt/gradle \
    JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64 \
    GOPATH=/go \
    GOTOOLCHAIN=local \
    GOMAXPROCS=4 \
    PATH=/opt/android-sdk/cmdline-tools/latest/bin:/opt/android-sdk/platform-tools:/opt/android-sdk/build-tools/35.0.0:/opt/gradle/bin:/go/bin:/usr/local/go/bin:$PATH

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates curl unzip openjdk-17-jdk-headless git build-essential python3 \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p "$ANDROID_HOME/cmdline-tools" /opt/rich2-toolchain \
    && curl --retry 3 --connect-timeout 30 -fsSL "$ANDROID_CMDLINE_TOOLS_URL" -o /tmp/android-cmdline.zip \
    && echo "$ANDROID_CMDLINE_TOOLS_SHA256  /tmp/android-cmdline.zip" | sha256sum -c - \
    && unzip -q /tmp/android-cmdline.zip -d /tmp/android-cmdline \
    && mv /tmp/android-cmdline/cmdline-tools "$ANDROID_HOME/cmdline-tools/latest" \
    && rm -rf /tmp/android-cmdline.zip /tmp/android-cmdline \
    && yes | sdkmanager --licenses >/dev/null \
    && sdkmanager --install \
        'platform-tools' 'platforms;android-35' 'build-tools;35.0.0' 'ndk;27.2.12479018' \
    && sdkmanager --list_installed > /opt/rich2-toolchain/sdk-packages.txt \
    && curl --retry 3 --connect-timeout 30 -fsSL "$GRADLE_URL" -o /tmp/gradle.zip \
    && echo "$GRADLE_SHA256  /tmp/gradle.zip" | sha256sum -c - \
    && unzip -q /tmp/gradle.zip -d /opt \
    && mv /opt/gradle-8.9 "$GRADLE_HOME" \
    && rm -f /tmp/gradle.zip

COPY --from=golang:1.24-bookworm@sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac /usr/local/go /usr/local/go-1.24
RUN rm -rf /usr/local/go && mv /usr/local/go-1.24 /usr/local/go
ENV GOLANG_VERSION=1.24.13

RUN go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@v2.8.8 \
    && go version > /opt/rich2-toolchain/go-version.txt \
    && go version -m /go/bin/ebitenmobile > /opt/rich2-toolchain/ebitenmobile-buildinfo.txt \
    && java -version 2> /opt/rich2-toolchain/java-version.txt \
    && gradle --version > /opt/rich2-toolchain/gradle-version.txt \
    && find "$ANDROID_HOME" -name source.properties -type f -exec sha256sum {} + \
       > /opt/rich2-toolchain/sdk-source-properties.sha256 \
    && chmod -R a+rX "$ANDROID_HOME" "$GRADLE_HOME" /go /opt/rich2-toolchain

LABEL org.opencontainers.image.title="logh3 Android build tools" \
      org.opencontainers.image.version="20261119-r1" \
      org.opencontainers.image.description="Go1.24.13 Ebiten2.8.8 SDK35 NDK27.2 Gradle8.9; no original assets"

USER 1000:1000
WORKDIR /src
CMD ["bash"]
