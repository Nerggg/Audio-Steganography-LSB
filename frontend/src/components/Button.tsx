"use client"

import type React from "react"
import type { ButtonProps } from "../types"

const Button: React.FC<ButtonProps> = ({
  children,
  onClick,
  disabled = false,
  variant = "primary",
  type = "button",
  className = "",
}) => {
  const baseClasses =
    "px-6 py-3 font-semibold rounded-lg transition-all duration-300 transform relative overflow-hidden group"

  const variantClasses = {
    primary: disabled
      ? "bg-gray-600 text-gray-400 cursor-not-allowed"
      : "bg-gradient-to-r from-pink-500 to-cyan-400 text-black hover:from-cyan-400 hover:to-pink-500 hover:scale-105 shadow-lg shadow-pink-500/25 hover:shadow-cyan-400/25 neon-border",
    secondary: disabled
      ? "border border-gray-600 text-gray-400 cursor-not-allowed"
      : "border border-cyan-400 text-cyan-400 hover:bg-cyan-400 hover:text-black hover:scale-105 hover:shadow-lg hover:shadow-cyan-400/25",
  }

  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      className={`${baseClasses} ${variantClasses[variant]} ${className}`}
    >
      {/* Animated background effect */}
      {!disabled && (
        <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover:translate-x-full transition-transform duration-700"></div>
      )}

      {/* Button content */}
      <span className="relative z-10">{children}</span>

      {/* Corner accents */}
      {!disabled && (
        <>
          <div className="absolute top-0 left-0 w-2 h-2 border-l-2 border-t-2 border-current opacity-50"></div>
          <div className="absolute top-0 right-0 w-2 h-2 border-r-2 border-t-2 border-current opacity-50"></div>
          <div className="absolute bottom-0 left-0 w-2 h-2 border-l-2 border-b-2 border-current opacity-50"></div>
          <div className="absolute bottom-0 right-0 w-2 h-2 border-r-2 border-b-2 border-current opacity-50"></div>
        </>
      )}
    </button>
  )
}

export default Button
