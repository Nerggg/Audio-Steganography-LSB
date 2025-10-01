import type React from "react"
// TypeScript interfaces and types for the audio steganography application

export interface UploadedFile {
  file: File
  name: string
  size: number
  url: string
}

export interface AppStatus {
  isLoading: boolean
  message: string
  type: "info" | "success" | "error" | "warning"
}

// Steganography method types
export type SteganographyMethod = 'lsb' | 'parity'

export interface EmbedOptions {
  stegKey: string
  method: SteganographyMethod
  nLSB: 1 | 2 | 3 | 4  // Only used for LSB method
  encrypt: boolean
  randomStart: boolean
}

export interface ExtractOptions {
  stegKey?: string
  method?: SteganographyMethod  // Optional: auto-detect if not provided
}

export interface EmbedResult {
  success: boolean
  psnr?: number
  stegoAudioUrl?: string
  message?: string
  secretSize?: number
  processingTime?: number
  embeddingMethod?: string
}

export interface ExtractResult {
  success: boolean
  message?: string
  error?: string
}

// Updated capacity interface to include both LSB and Parity capacities
export interface AudioCapacity {
  "1_lsb": number
  "2_lsb": number
  "3_lsb": number
  "4_lsb": number
  "parity": number
}

// Method information for UI display
export interface MethodInfo {
  id: SteganographyMethod
  name: string
  description: string
  advantages: string[]
  disadvantages: string[]
  capacity: 'high' | 'medium' | 'low'
  robustness: 'high' | 'medium' | 'low'
}

// Enhanced status display with capacity breakdown
export interface CapacityInfo {
  audioCapacity: AudioCapacity
  selectedMethod: SteganographyMethod
  selectedLSB?: number
  isCapacitySufficient: boolean
  requiredBytes: number
  availableBytes: number
}

export type ButtonVariant = "primary" | "secondary"

export interface ButtonProps {
  children: React.ReactNode
  onClick?: () => void
  disabled?: boolean
  variant?: ButtonVariant
  type?: "button" | "submit" | "reset"
  className?: string
}

export interface FileUploadProps {
  onFileSelect: (file: UploadedFile) => void
  accept?: string
  label: string
  currentFile?: UploadedFile
}

export interface AudioPlayerProps {
  audioUrl?: string
  label?: string
  className?: string
  onClick?: () => void
}

export interface StatusDisplayProps {
  status: AppStatus
  capacityInfo?: CapacityInfo
  psnr?: number
}
