import type React from "react"

interface TooltipProps {
  content: string
  children: React.ReactNode
  className?: string
}

export const Tooltip: React.FC<TooltipProps> = ({ content, children, className = "" }) => {
  return (
    <div className={`group relative inline-block ${className}`}>
      {children}
      <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-3 py-2 bg-gray-900 text-white text-sm rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none whitespace-nowrap z-10 border border-gray-700">
        {content}
        <div className="absolute top-full left-1/2 transform -translate-x-1/2 border-4 border-transparent border-t-gray-900"></div>
      </div>
    </div>
  )
}

interface InfoIconProps {
  tooltip: string
  className?: string
}

export const InfoIcon: React.FC<InfoIconProps> = ({ tooltip, className = "" }) => {
  return (
    <Tooltip content={tooltip} className={className}>
      <div className="inline-flex items-center justify-center w-4 h-4 rounded-full bg-blue-500/20 text-blue-400 text-xs cursor-help border border-blue-500/30 hover:bg-blue-500/30 transition-colors">
        ?
      </div>
    </Tooltip>
  )
}

export const MethodComparison: React.FC = () => {
  return (
    <div className="grid grid-cols-2 gap-4 text-xs">
      <div className="space-y-2">
        <div className="font-semibold text-blue-300">LSB Method</div>
        <div className="space-y-1 text-gray-400">
          <div>• High capacity (1-4x)</div>
          <div>• Configurable embedding</div>
          <div>• Fast processing</div>
          <div>• Good for large files</div>
        </div>
        <div className="text-yellow-400 text-xs">⚠ Less robust to compression</div>
      </div>
      
      <div className="space-y-2">
        <div className="font-semibold text-green-300">Parity Method</div>
        <div className="space-y-1 text-gray-400">
          <div>• High robustness</div>
          <div>• Compression resistant</div>
          <div>• Minimal artifacts</div>
          <div>• Good for critical files</div>
        </div>
        <div className="text-yellow-400 text-xs">⚠ Lower capacity (1 bit/byte)</div>
      </div>
    </div>
  )
}

// Enhanced error message component
interface ErrorMessageProps {
  title: string
  message: string
  suggestions?: string[]
  onRetry?: () => void
}

export const ErrorMessage: React.FC<ErrorMessageProps> = ({ title, message, suggestions, onRetry }) => {
  return (
    <div className="border border-red-400/50 rounded-lg p-4 bg-red-900/10 backdrop-blur-sm">
      <div className="flex items-center gap-2 mb-3">
        <div className="w-6 h-6 rounded-full bg-red-500/20 text-red-400 flex items-center justify-center text-sm">
          ✗
        </div>
        <h3 className="text-red-400 font-mono font-bold">{title}</h3>
      </div>
      
      <div className="text-red-300 mb-3">{message}</div>
      
      {suggestions && suggestions.length > 0 && (
        <div className="mb-3">
          <div className="text-yellow-400 text-sm font-semibold mb-2">Suggestions:</div>
          <ul className="space-y-1 text-sm text-gray-300">
            {suggestions.map((suggestion, index) => (
              <li key={index}>• {suggestion}</li>
            ))}
          </ul>
        </div>
      )}
      
      {onRetry && (
        <button
          onClick={onRetry}
          className="px-4 py-2 bg-red-500/20 border border-red-400/50 text-red-300 rounded hover:bg-red-500/30 transition-colors text-sm font-mono"
        >
          TRY AGAIN
        </button>
      )}
    </div>
  )
}

// Success message with additional info
interface SuccessMessageProps {
  title: string
  message: string
  details?: { label: string; value: string }[]
}

export const SuccessMessage: React.FC<SuccessMessageProps> = ({ title, message, details }) => {
  return (
    <div className="border border-green-400/50 rounded-lg p-4 bg-green-900/10 backdrop-blur-sm">
      <div className="flex items-center gap-2 mb-3">
        <div className="w-6 h-6 rounded-full bg-green-500/20 text-green-400 flex items-center justify-center text-sm">
          ✓
        </div>
        <h3 className="text-green-400 font-mono font-bold">{title}</h3>
      </div>
      
      <div className="text-green-300 mb-3">{message}</div>
      
      {details && details.length > 0 && (
        <div className="grid grid-cols-2 gap-4">
          {details.map((detail, index) => (
            <div key={index} className="text-sm">
              <span className="text-gray-400">{detail.label}:</span>
              <span className="text-white ml-2 font-mono">{detail.value}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}