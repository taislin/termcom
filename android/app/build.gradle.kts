plugins {
    id("com.android.application")
}

android {
    namespace = "com.taislin.termcom"
    compileSdk = 34

    defaultConfig {
        applicationId = "com.taislin.termcom"
        minSdk = 21
        targetSdk = 34
        versionCode = 473
        versionName = "0.47.3"

        ndk {
            abiFilters += listOf("arm64-v8a")
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = false
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro")
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_1_8
        targetCompatibility = JavaVersion.VERSION_1_8
    }
}

dependencies {
    implementation(fileTree(mapOf("dir" to "libs", "include" to listOf("*.aar"))))
}
