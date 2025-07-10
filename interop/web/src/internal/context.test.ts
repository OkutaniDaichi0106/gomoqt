import { 
    Context, 
    background, 
    withSignal, 
    withCancel, 
    withCancelCause, 
    withTimeout, 
    withPromise 
} from './context';

describe('Context', () => {
    describe('background', () => {
        it('should create a background context', () => {
            const ctx = background();
            expect(ctx).toBeDefined();
            expect(ctx.signal).toBeInstanceOf(AbortSignal);
            expect(ctx.err()).toBeNull();
        });

        it('should return the same instance on multiple calls', () => {
            const ctx1 = background();
            const ctx2 = background();
            expect(ctx1).toBe(ctx2);
        });

        it('should not be aborted initially', () => {
            const ctx = background();
            expect(ctx.signal.aborted).toBe(false);
            expect(ctx.err()).toBeNull();
        });
    });

    describe('withSignal', () => {
        it('should create a context with custom signal', () => {
            const parentCtx = background();
            const controller = new AbortController();
            const childCtx = withSignal(parentCtx, controller.signal);
            
            expect(childCtx.signal).toBeDefined();
            expect(childCtx.err()).toBeNull();
        });

        it('should be cancelled when custom signal is aborted', async () => {
            const parentCtx = background();
            const controller = new AbortController();
            const childCtx = withSignal(parentCtx, controller.signal);
            
            let done = false;
            childCtx.done().catch(() => { done = true; });
            
            controller.abort(new Error('Custom abort'));
            
            // Wait a bit for async operations
            await new Promise(resolve => setTimeout(resolve, 10));
            
            expect(done).toBe(true);
            expect(childCtx.signal.aborted).toBe(true);
            expect(childCtx.err()).toBeInstanceOf(Error);
        });

        it('should handle already aborted signal', () => {
            const parentCtx = background();
            const controller = new AbortController();
            const testError = new Error('Already aborted');
            controller.abort(testError);
            
            const childCtx = withSignal(parentCtx, controller.signal);
            expect(childCtx.signal.aborted).toBe(true);
            expect(childCtx.err()).toBeInstanceOf(Error);
        });
    });

    describe('withCancel', () => {
        it('should create a cancellable context', () => {
            const parentCtx = background();
            const [childCtx, cancel] = withCancel(parentCtx);
            
            expect(childCtx.signal).toBeDefined();
            expect(childCtx.err()).toBeNull();
            expect(typeof cancel).toBe('function');
        });

        it('should cancel the context when cancel function is called', async () => {
            const parentCtx = background();
            const [childCtx, cancel] = withCancel(parentCtx);
            
            let done = false;
            childCtx.done().catch(() => { done = true; });
            
            cancel();
            
            // Wait a bit for async operations
            await new Promise(resolve => setTimeout(resolve, 10));
            
            expect(done).toBe(true);
            expect(childCtx.signal.aborted).toBe(true);
            expect(childCtx.err()).toBeInstanceOf(Error);
        });

        it('should be cancelled when parent is cancelled', async () => {
            const [parentCtx, parentCancel] = withCancel(background());
            const [childCtx, ] = withCancel(parentCtx);
            
            let childDone = false;
            childCtx.done().catch(() => { childDone = true; });
            
            parentCancel();
            
            // Wait a bit for async operations
            await new Promise(resolve => setTimeout(resolve, 10));
            
            expect(childDone).toBe(true);
            expect(childCtx.signal.aborted).toBe(true);
        });
    });

    describe('withCancelCause', () => {
        it('should create a cancellable context with custom error', () => {
            const parentCtx = background();
            const [childCtx, cancelWithCause] = withCancelCause(parentCtx);
            
            expect(childCtx.signal).toBeDefined();
            expect(childCtx.err()).toBeNull();
            expect(typeof cancelWithCause).toBe('function');
        });

        it('should cancel with custom error', async () => {
            const parentCtx = background();
            const [childCtx, cancelWithCause] = withCancelCause(parentCtx);
            
            const customError = new Error('Custom cancellation');
            let done = false;
            
            childCtx.done().catch((err) => { 
                done = true; 
                expect(err).toBe(customError);
            });
            
            cancelWithCause(customError);
            
            // Wait a bit for async operations
            await new Promise(resolve => setTimeout(resolve, 10));
            
            expect(done).toBe(true);
            expect(childCtx.signal.aborted).toBe(true);
        });

        it('should handle null error', async () => {
            const parentCtx = background();
            const [childCtx, cancelWithCause] = withCancelCause(parentCtx);
            
            let done = false;
            childCtx.done().catch(() => { done = true; });
            
            cancelWithCause(null);
            
            // Wait a bit for async operations
            await new Promise(resolve => setTimeout(resolve, 10));
            
            expect(done).toBe(true);
            expect(childCtx.signal.aborted).toBe(true);
        });
    });

    describe('withTimeout', () => {
        it('should create a context with timeout', () => {
            const parentCtx = background();
            const childCtx = withTimeout(parentCtx, 1000);
            
            expect(childCtx.signal).toBeDefined();
            expect(childCtx.err()).toBeNull();
        });

        it('should cancel after timeout', async () => {
            const parentCtx = background();
            const childCtx = withTimeout(parentCtx, 50);
            
            let done = false;
            let caughtError: any = null;
            
            childCtx.done().catch((err) => { 
                done = true; 
                caughtError = err;
            });
            
            // Wait longer than timeout
            await new Promise(resolve => setTimeout(resolve, 100));
            
            expect(done).toBe(true);
            expect(childCtx.signal.aborted).toBe(true);
            expect(caughtError).toBeInstanceOf(Error);
            if (caughtError instanceof Error) {
                expect(caughtError.message).toContain('timeout');
            }
        });

        it('should not timeout if parent is cancelled first', async () => {
            const [parentCtx, parentCancel] = withCancel(background());
            const childCtx = withTimeout(parentCtx, 1000);
            
            let done = false;
            childCtx.done().catch(() => { done = true; });
            
            // Cancel parent before timeout
            parentCancel();
            
            // Wait a bit for async operations
            await new Promise(resolve => setTimeout(resolve, 10));
            
            expect(done).toBe(true);
            expect(childCtx.signal.aborted).toBe(true);
        });
    });

    describe('withPromise', () => {
        it('should create a context that cancels when promise resolves', async () => {
            const parentCtx = background();
            const promise = Promise.resolve('test value');
            const childCtx = withPromise(parentCtx, promise);
            
            let done = false;
            childCtx.done().catch(() => { done = true; });
            
            // Wait for promise to resolve
            await new Promise(resolve => setTimeout(resolve, 10));
            
            expect(done).toBe(true);
            expect(childCtx.signal.aborted).toBe(true);
        });

        it('should create a context that cancels when promise rejects', async () => {
            const parentCtx = background();
            const testError = new Error('Promise rejected');
            const promise = Promise.reject(testError);
            const childCtx = withPromise(parentCtx, promise);
            
            let done = false;
            let error: Error | null = null;
            
            childCtx.done().catch((err) => { 
                done = true; 
                error = err;
            });
            
            // Wait for promise to reject
            await new Promise(resolve => setTimeout(resolve, 10));
            
            expect(done).toBe(true);
            expect(childCtx.signal.aborted).toBe(true);
            expect(error).toBe(testError);
        });

        it('should handle non-Error rejection reasons', async () => {
            const parentCtx = background();
            const promise = Promise.reject('string rejection');
            const childCtx = withPromise(parentCtx, promise);
            
            let done = false;
            let caughtError: any = null;
            
            childCtx.done().catch((err) => { 
                done = true; 
                caughtError = err;
            });
            
            // Wait for promise to reject
            await new Promise(resolve => setTimeout(resolve, 10));
            
            expect(done).toBe(true);
            expect(childCtx.signal.aborted).toBe(true);
            expect(caughtError).toBeInstanceOf(Error);
            if (caughtError instanceof Error) {
                expect(caughtError.message).toContain('string rejection');
            }
        });
    });

    describe('Context interface', () => {
        it('should provide done() promise that rejects when cancelled', async () => {
            const [ctx, cancel] = withCancel(background());
            
            let rejected = false;
            let error: Error | null = null;
            
            ctx.done().catch((err) => {
                rejected = true;
                error = err;
            });
            
            cancel();
            
            // Wait a bit for async operations
            await new Promise(resolve => setTimeout(resolve, 10));
            
            expect(rejected).toBe(true);
            expect(error).toBeInstanceOf(Error);
        });

        it('should have signal property', () => {
            const ctx = background();
            expect(ctx.signal).toBeInstanceOf(AbortSignal);
        });

        it('should return error when cancelled', async () => {
            const [ctx, cancel] = withCancel(background());
            
            expect(ctx.err()).toBeNull();
            
            cancel();
            
            // Wait a bit for async operations
            await new Promise(resolve => setTimeout(resolve, 10));
            
            expect(ctx.err()).toBeInstanceOf(Error);
        });
    });
});
