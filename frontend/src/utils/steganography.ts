import type { MethodInfo, SteganographyMethod } from '../types'

// Steganography method information for UI display
export const STEGANOGRAPHY_METHODS: Record<SteganographyMethod, MethodInfo> = {
  lsb: {
    id: 'lsb',
    name: 'LSB Method',
    description: 'Least Significant Bit steganography modifies the least significant bits of audio samples.',
    advantages: [
      'High embedding capacity',
      'Configurable capacity (1-4 LSBs)',
      'Fast embedding/extraction',
      'Good for large files'
    ],
    disadvantages: [
      'More susceptible to compression',
      'Less robust against noise',
      'May cause audible artifacts at high LSB counts'
    ],
    capacity: 'high',
    robustness: 'medium'
  },
  parity: {
    id: 'parity',
    name: 'Parity Method',
    description: 'Parity steganography embeds data by adjusting the parity (even/odd bit count) of audio samples.',
    advantages: [
      'High robustness against compression',
      'Resistant to noise',
      'Minimal audible artifacts',
      'Good for small, critical files'
    ],
    disadvantages: [
      'Lower embedding capacity',
      'Fixed capacity (1 bit per byte)',
      'Slower for large files'
    ],
    capacity: 'low',
    robustness: 'high'
  }
}

// PSNR quality thresholds (as per specification)
export const PSNR_THRESHOLDS = {
  MINIMUM: 30, // Minimum acceptable PSNR (30 dB as per specification)
  EXCELLENT: 40 // Excellent quality threshold
} as const

// Default embed options
export const DEFAULT_EMBED_OPTIONS = {
  stegKey: '',
  method: 'lsb' as SteganographyMethod,
  nLSB: 1 as 1 | 2 | 3 | 4,
  encrypt: false,
  randomStart: false
}

// PSNR quality assessment function
export const assessPSNRQuality = (psnr: number): {
  level: 'excellent' | 'good' | 'acceptable' | 'poor'
  description: string
  color: string
} => {
  if (psnr >= PSNR_THRESHOLDS.EXCELLENT) {
    return {
      level: 'excellent',
      description: 'Excellent audio quality',
      color: 'green'
    }
  } else if (psnr >= PSNR_THRESHOLDS.MINIMUM) {
    return {
      level: 'good', 
      description: 'Good audio quality preserved',
      color: 'green'
    }
  } else if (psnr >= 20) {
    return {
      level: 'acceptable',
      description: 'Acceptable quality, slight degradation',
      color: 'yellow'
    }
  } else {
    return {
      level: 'poor',
      description: 'Poor quality, significant degradation',
      color: 'red'
    }
  }
}

// Default extract options
export const DEFAULT_EXTRACT_OPTIONS = {
  stegKey: '',
  method: undefined as SteganographyMethod | undefined
}

// Helper functions
export const getCapacityForMethod = (
  capacities: Record<string, number>,
  method: SteganographyMethod,
  lsb?: number
): number => {
  if (method === 'parity') {
    return capacities.parity || 0
  }
  
  if (method === 'lsb' && lsb) {
    return capacities[`${lsb}_lsb`] || 0
  }
  
  return 0
}

export const formatCapacity = (bytes: number): string => {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export const getMethodCapacityDescription = (method: SteganographyMethod, lsb?: number): string => {
  if (method === 'parity') {
    return '1 bit per audio byte (most robust)'
  }
  
  if (method === 'lsb' && lsb) {
    return `${lsb} bit${lsb > 1 ? 's' : ''} per audio byte (${lsb}x capacity)`
  }
  
  return 'Variable capacity based on configuration'
}