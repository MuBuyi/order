import axios from 'axios'

let _cache = null

export async function fetchCountries() {
  if (_cache) return _cache
  try {
    const res = await axios.get('/api/countries')
    const items = res.data?.items || []
    _cache = items
    return items
  } catch (e) {
    // 失败时返回常用默认列表，避免前端空白
    return ['菲律宾', '印尼', '马来西亚']
  }
}

export function clearCountriesCache() {
  _cache = null
}
