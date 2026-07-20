// Lightweight API response shape validation.
// Warns (via console.warn) in development mode when a response doesn't match
// the expected schema. Never throws — purely diagnostic.

const isDev = typeof import.meta !== 'undefined' && import.meta.env?.DEV

/* ------------------------------------------------------------------ */
/*  Schema helpers                                                     */
/* ------------------------------------------------------------------ */

/**
 * Create a schema definition for an object-shaped response.
 *
 * @param {string[]}   required  – field names that MUST be present
 * @param {string[]}   [optional] – field names that MAY be present
 * @param {Object}     [types]   – field → expected typeof string, e.g. { id: 'number' }
 * @returns {object}   schema
 */
export function schema(required, optional = [], types = {}) {
  const all = [...required, ...optional]
  return { required, optional, types, all }
}

// Pre‑defined schemas for this app's API endpoints.

export const EVENT_SCHEMA = schema(
  ['id', 'unique_hash', 'city', 'address', 'start_at', 'end_at'],
  ['created_at'],
  {
    id: 'number',
    unique_hash: 'string',
    city: 'string',
    address: 'string',
    start_at: 'string',
    end_at: 'string',
    created_at: 'string',
  },
)

export const UPDATED_AT_SCHEMA = schema(
  ['created_at'],
  [],
  { created_at: 'string' },
)

/* ------------------------------------------------------------------ */
/*  Validation logic                                                   */
/* ------------------------------------------------------------------ */

/**
 * Validate a single object against a schema.
 * Logs warnings for missing required fields / unexpected fields / type mismatches.
 *
 * @param {object}  obj    – the parsed JSON object to check
 * @param {object}  schema – a schema created with `schema()`
 * @param {string}  label  – human‑readable label for warning messages
 * @returns {boolean}      – true if any problems were found
 */
export function validateShape(obj, schema, label) {
  if (!isDev) return false
  if (!obj || typeof obj !== 'object' || Array.isArray(obj)) {
    console.warn(`[validateApi] ${label}: expected a plain object, got`, obj)
    return true
  }

  let problems = 0
  const keys = Object.keys(obj)

  // Missing required fields
  for (const key of schema.required) {
    if (!(key in obj)) {
      console.warn(`[validateApi] ${label}: missing required field "${key}"`)
      problems++
    }
  }

  // Type mismatches for known fields that are present
  for (const [key, expectedType] of Object.entries(schema.types)) {
    if (key in obj && obj[key] !== null && obj[key] !== undefined) {
      const actual = typeof obj[key]
      if (actual !== expectedType) {
        console.warn(
          `[validateApi] ${label}: field "${key}" expected ${expectedType}, got ${actual} (value: ${JSON.stringify(obj[key]).slice(0, 80)})`,
        )
        problems++
      }
    }
  }

  // Unexpected fields (present in obj but not in schema)
  for (const key of keys) {
    if (!schema.all.includes(key)) {
      console.warn(`[validateApi] ${label}: unexpected field "${key}"`)
      problems++
    }
  }

  return problems > 0
}

/**
 * Validate an array of objects against a schema.
 * Samples up to `maxSample` items to avoid flooding the console.
 *
 * @param {object[]} arr         – the array to validate
 * @param {object}   schema      – a schema created with `schema()`
 * @param {string}   label       – label for warnings
 * @param {number}   [maxSample] – max items to inspect (default 3)
 * @returns {boolean} – true if any problems were found
 */
export function validateList(arr, schema, label, maxSample = 3) {
  if (!isDev) return false
  if (!Array.isArray(arr)) {
    console.warn(`[validateApi] ${label}: expected an array, got`, arr)
    return true
  }

  let problems = 0
  const sample = arr.slice(0, maxSample)
  for (let i = 0; i < sample.length; i++) {      if (validateShape(sample[i], schema, `${label}[${i}]`)) problems++
  }

  // If the array has items but our sample was empty / all null, flag it
  if (arr.length > 0 && sample.length === 0) {
    console.warn(`[validateApi] ${label}: array has ${arr.length} items, but none could be sampled`)
    problems++
  }

  return problems > 0
}

/**
 * Convenience: validate a response value that could be an array or a single object.
 * Calls validateList for arrays and validateShape for objects.
 *
 * @param {object|object[]} data
 * @param {object}          schema
 * @param {string}          label
 * @returns {boolean}
 */
export function validateResponse(data, schema, label) {
  if (Array.isArray(data)) return validateList(data, schema, label)
  return validateShape(data, schema, label)
}
