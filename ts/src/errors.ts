export class MountOSError extends Error {
  readonly status: number
  readonly errorCode?: number

  constructor(message: string, status: number, errorCode?: number) {
    super(message)
    this.name = 'MountOSError'
    this.status = status
    this.errorCode = errorCode
  }
}
