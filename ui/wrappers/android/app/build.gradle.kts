import java.util.Properties

plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
}

// Load local.properties (like .env for Android)
val localProps = Properties().also { props ->
    val f = rootProject.file("local.properties")
    if (f.exists()) props.load(f.inputStream())
}

fun localProp(key: String, default: String): String =
    (localProps[key] as? String)?.takeIf { it.isNotBlank() } ?: default

android {
    namespace = "com.timepad.android"
    compileSdk = 34

    defaultConfig {
        applicationId = "com.timepad.android"
        minSdk = 26
        targetSdk = 34
        versionCode = 1
        versionName = "1.0"

        // Expose env-style config to BuildConfig
        buildConfigField(
            "String", "SERVER_URL",
            "\"${localProp("TIMEPAD_SERVER_URL", "http://10.0.2.2:8080/api/v1")}\""
        )
        buildConfigField(
            "String", "DASHBOARD_URL",
            "\"${localProp("TIMEPAD_DASHBOARD_URL", "http://10.0.2.2:5173")}\""
        )
    }

    buildFeatures {
        buildConfig = true
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }
}

dependencies {
    implementation("androidx.core:core-ktx:1.13.1")
    implementation("androidx.appcompat:appcompat:1.7.0")
    implementation("com.squareup.okhttp3:okhttp:4.12.0")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.8.1")
}
