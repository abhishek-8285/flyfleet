package com.flyfleet.vehicledevice.retention

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import java.nio.file.Path

data class RetentionConfigSource(
    val criticalRetentionDays: Int,
    val evidenceReserveMb: Long,
    val legalHold: Boolean,
)

@Serializable
private data class RetentionConfigDto(
    val criticalRetentionDays: Int = 30,
    val evidenceReserveMb: Long = 5000,
    @SerialName("legal_hold")
    val legalHold: Boolean = false,
)

class RetentionConfigProvider(private val configPath: Path) {
    fun load(): RetentionConfig {
        val text = configPath.toFile().readText()
        val json = Json.parseToJsonElement(text).jsonObject
        val retention = json["retention"]?.jsonObject ?: Json.parseToJsonElement("{}").jsonObject
        val dto = Json.decodeFromJsonElement(RetentionConfigDto.serializer(), retention)
        return RetentionConfig(
            criticalRetentionDays = dto.criticalRetentionDays,
            evidenceReserveMb = dto.evidenceReserveMb,
            legalHold = dto.legalHold,
        )
    }
}
