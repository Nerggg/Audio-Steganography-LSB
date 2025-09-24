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

export interface EmbedOptions {
  stegKey: string
  nLSB: 1 | 2 | 4
  encrypt: boolean
  randomStart: boolean
}

export interface ExtractOptions {
  stegKey?: string
}

export interface EmbedResult {
  success: boolean
  psnr?: number
  stegoAudioUrl?: string
  message?: string
}

export interface ExtractResult {
  success: boolean
  message?: string
  error?: string
}

export interface AudioCapacity {
  maxBytes: number
  maxCharacters: number
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
}

export interface StatusDisplayProps {
  status: AppStatus
  capacity?: AudioCapacity
  psnr?: number
}
