import { ComponentType, ErrorInfo, FC, memo, PureComponent, ReactNode } from "react"
import flow from "lodash.flow"

export const NullFallback: FC = () => null

export type ErrorBoundaryProps = {
  Fallback: ReactNode
}

export type ErrorBoundaryState = {
  error?: Error
  errorInfo?: ErrorInfo
}

export class ErrorBoundary extends PureComponent<
  ErrorBoundaryProps,
  ErrorBoundaryState
> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = { error: undefined, errorInfo: undefined }
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    this.setState({
      error: error,
      errorInfo: errorInfo,
    })
    /**
     * @todo log Sentry here
     */
  }

  public render(): ReactNode {
    if (this.state.error) return this.props.Fallback
    return this.props.children
  }
}

export function withErrorBoundary<Props>(Component: ComponentType<Props>) {
  return (Fallback = NullFallback) => {
    // eslint-disable-next-line react/display-name
    return memo(({ ...props }: Props) => {
      return (
        <ErrorBoundary Fallback={<Fallback {...props} />}>
          <Component {...props} />
        </ErrorBoundary>
      )
    })
  }
}

export const enhance: <Props>(
  Component: FC<Props>
) => (
  Fallback?: FC
) => React.MemoExoticComponent<({ ...props }: Props) => JSX.Element> = flow(
  memo,
  withErrorBoundary
)

export type TcResult<Data, Throws = Error> = [null, Data] | [Throws]

export async function tcAsync<T, Throws = Error>(
  promise: Promise<T>
): Promise<TcResult<T, Throws>> {
  try {
    const response: T = await promise

    return [null, response]
  } catch (error) {
    return [error] as [Throws]
  }
}

export function tcSync<
  ArrType,
  Params extends Array<ArrType>,
  Returns,
  Throws = Error
>(
  fn: (...params: Params) => Returns,
  ...deps: Params
): TcResult<Returns, Throws> {
  try {
    const data: Returns = fn(...deps)

    return [null, data]
  } catch (e) {
    return [e] as [Throws]
  }
}

export function exhaustiveCheck(x: never): void {
  throw new Error(`${x} should be unreachable`)
}