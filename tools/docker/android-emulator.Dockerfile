# logh3 API35 模擬器（規格161）。下載網址與 SHA-256 沿用 ~/cht/rich2/tools/docker/android-emulator.Dockerfile
# 已核對的版本（emulator 37.2.13、x86_64 API35 r02），建置時照樣逐檔驗證。
# 與 rich2 不同：執行時帶 --device /dev/kvm 用硬體加速（主機有 KVM）。不含任何遊戲資料。
FROM logh3-android:20261119-r1

USER root
ARG EMULATOR_URL=https://dl.google.com/android/repository/emulator-linux_x64-16740729.zip
ARG EMULATOR_SHA256=cbc7c3cfb83220b8a0f9c513fdab557712e74d339ec19ddffa43f23350ae9bdd
ARG SYSTEM_IMAGE_URL=https://dl.google.com/android/repository/sys-img/android/x86_64-35_r02.zip
ARG SYSTEM_IMAGE_SHA256=6dd7de33e63ef105cf2fabea6badda1dbe7665c96d8908e5f6e1407e63ff4556

ENV PATH=/opt/android-sdk/emulator:$PATH \
    QT_QPA_PLATFORM=offscreen \
    HOME=/out/home \
    ANDROID_AVD_HOME=/out/avd \
    ANDROID_USER_HOME=/out/home/.android \
    ANDROID_EMULATOR_HOME=/out/home/.android

RUN apt-get update && apt-get install -y --no-install-recommends \
        libgl1 libnss3 libx11-6 libxcb1 libxcb-cursor0 libpulse0 libasound2 libfontconfig1 \
        libxcomposite1 libxcursor1 libxi6 libxrandr2 libxtst6 libxdamage1 libxfixes3 \
        libxkbcommon0 libxkbcommon-x11-0 libdbus-1-3 libgbm1 libdrm2 libpcre2-16-0 \
    && rm -rf /var/lib/apt/lists/*

RUN apt-get update && apt-get install -y --no-install-recommends libxkbfile1 \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /opt/logh3-emulator-toolchain "$ANDROID_HOME/system-images/android-35/default" \
    && curl --retry 3 --connect-timeout 30 -fsSL "$EMULATOR_URL" -o /tmp/emulator.zip \
    && echo "$EMULATOR_SHA256  /tmp/emulator.zip" | sha256sum -c - \
    && curl --retry 3 --connect-timeout 30 -fsSL "$SYSTEM_IMAGE_URL" -o /tmp/system-image.zip \
    && echo "$SYSTEM_IMAGE_SHA256  /tmp/system-image.zip" | sha256sum -c - \
    && printf '%s  %s\n%s  %s\n' "$EMULATOR_SHA256" "$EMULATOR_URL" "$SYSTEM_IMAGE_SHA256" "$SYSTEM_IMAGE_URL" \
       > /opt/logh3-emulator-toolchain/archives.sha256 \
    && unzip -q /tmp/emulator.zip -d "$ANDROID_HOME" \
    && unzip -q /tmp/system-image.zip -d "$ANDROID_HOME/system-images/android-35/default" \
    && rm -f /tmp/emulator.zip /tmp/system-image.zip \
    && sdkmanager --list_installed > /opt/logh3-emulator-toolchain/sdk-packages.txt \
    && (emulator -version > /opt/logh3-emulator-toolchain/emulator-version.txt 2>&1 \
        || { cat /opt/logh3-emulator-toolchain/emulator-version.txt; exit 1; }) \
    && find "$ANDROID_HOME/emulator" "$ANDROID_HOME/system-images/android-35/default" \
       -name package.xml -type f -exec sha256sum {} + > /opt/logh3-emulator-toolchain/package-xml.sha256

# The emulator archive omits package.xml. Register its verified source.properties
# revision using the generic local-package schema already supplied by this SDK.
RUN python3 -c 'import os,xml.etree.ElementTree as E; from pathlib import Path; sdk=Path(os.environ["ANDROID_HOME"]); src=(sdk/"emulator/source.properties").read_text(); assert "Pkg.Revision=37.2.13" in src; tree=E.parse(sdk/"platform-tools/package.xml"); tree.getroot().set("xmlns:generic","http://schemas.android.com/repository/android/generic/02"); p=tree.getroot().find("localPackage"); p.set("path","emulator"); p.find("type-details").set("{http://www.w3.org/2001/XMLSchema-instance}type","generic:genericDetailsType"); p.find("display-name").text="Android Emulator"; r=p.find("revision"); r.clear(); [(E.SubElement(r,k).__setattr__("text",v)) for k,v in [("major","37"),("minor","2"),("micro","13")]]; tree.write(sdk/"emulator/package.xml",encoding="UTF-8",xml_declaration=True)' \
    && sdkmanager --list_installed > /opt/logh3-emulator-toolchain/sdk-packages.txt \
    && python3 -c 'from pathlib import Path; import re; s=Path("/opt/logh3-emulator-toolchain/sdk-packages.txt").read_text(); assert re.search(r"emulator\s+\|\s+37\.2\.13\s+\|",s),s' \
    && find "$ANDROID_HOME/emulator" "$ANDROID_HOME/system-images/android-35/default" \
       -name package.xml -type f -exec sha256sum {} + > /opt/logh3-emulator-toolchain/package-xml.sha256

# Version probes run during image construction, but runtime is UID 1000.
# Do not bake a root-owned emulator crash-report directory into /tmp.
RUN rm -rf /tmp/android-unknown \
    && dpkg-query -W -f='${Package}\t${Version}\n' \
       > /opt/logh3-emulator-toolchain/runtime-package-versions.txt

LABEL org.opencontainers.image.title="logh3 API35 emulator" \
      org.opencontainers.image.version="api35-20261119-r1" \
      org.opencontainers.image.base.name="logh3-android:20261119-r1" \
      org.opencontainers.image.description="AOSP/default x86_64 API35 r2, emulator37.2.13; KVM at runtime; no game data"

USER 1000:1000
WORKDIR /out
CMD ["bash"]
