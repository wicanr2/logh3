/*
 * GIN3PS 資料模型：位元組版面常數與小端讀取。
 * 版面來源見 docs/internals/33（艦隊表）與 64（艦艇池）。
 */
#ifndef GIN3_MODEL_H
#define GIN3_MODEL_H

#include <stdint.h>

#define GIN3_FLEET_SIZE 122
#define GIN3_FLEET_SLOT_COUNT 0x0F
#define GIN3_FLEET_SLOTS 0x10
#define GIN3_FLEET_MAX_SLOTS 12
#define GIN3_FLEET_ABILITY40 0x40

#define GIN3_SHIP_SIZE 78
#define GIN3_SHIP_KIND 0x04
#define GIN3_SHIP_COUNT 0x08
#define GIN3_POOL_SHIPS 352

static inline uint16_t gin3_le16(const uint8_t *p)
{
    return (uint16_t)(p[0] | (p[1] << 8));
}

uint16_t gin3ps_fleet_attack(const uint8_t *fleet, const uint8_t *pool);

#endif
