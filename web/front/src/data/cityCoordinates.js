// Fallback coordinates for known cities (used when geocoding API fails or as a quick cache)
// All cities in Mazandaran province, Iran (IRST, +03:30).
export const CITY_COORDS = {
  'قایمشهر': { latitude: 36.4633, longitude: 52.8619, name_en: 'Qaemshahr' },
  'آمل': { latitude: 36.4669, longitude: 52.3497, name_en: 'Amol' },
  'بابل': { latitude: 36.55, longitude: 52.6833, name_en: 'Babol' },
  'سوادکوه': { latitude: 36.1833, longitude: 53.0333, name_en: 'Savadkuh' },
  'ساری': { latitude: 36.5633, longitude: 53.06, name_en: 'Sari' },
  'نوشهر': { latitude: 36.6483, longitude: 51.4961, name_en: 'Noshahr' },
  'چالوس': { latitude: 36.6553, longitude: 51.4206, name_en: 'Chalus' },
  'تنکابن': { latitude: 36.8164, longitude: 50.8731, name_en: 'Tonekabon' },
  'رامسر': { latitude: 36.9031, longitude: 50.6583, name_en: 'Ramsar' },
  'بهشهر': { latitude: 36.6942, longitude: 53.5411, name_en: 'Behshahr' },
  'گرگان': { latitude: 36.8395, longitude: 54.4353, name_en: 'Gorgan' },
  'محمودآباد': { latitude: 36.6314, longitude: 52.2656, name_en: 'Mahmudabad' },
  'فریدونکنار': { latitude: 36.6856, longitude: 52.5211, name_en: 'Fereydunkenar' },
  'نکا': { latitude: 36.6456, longitude: 53.2983, name_en: 'Neka' },
  'میاندورود': { latitude: 36.6081, longitude: 53.2131, name_en: 'Miandorud' },
  'سیمرغ': { latitude: 36.6092, longitude: 52.7319, name_en: 'Simreh' },
  'جویبار': { latitude: 36.6422, longitude: 52.9047, name_en: 'Juybar' },
  'کلاردشت': { latitude: 36.5164, longitude: 51.0981, name_en: 'Kalardasht' },
  'عباس‌آباد': { latitude: 36.7267, longitude: 51.1136, name_en: 'Abbasabad' },
  'علی‌آباد': { latitude: 36.8239, longitude: 54.8867, name_en: 'Aliabad' },
  'کردکوی': { latitude: 36.7975, longitude: 54.1117, name_en: 'Kordkuy' },
  'بندرگز': { latitude: 36.7756, longitude: 53.9481, name_en: 'Bandargaz' },
}

export function getCoords(city) {
  if (!city) return null
  return CITY_COORDS[city] || null
}
