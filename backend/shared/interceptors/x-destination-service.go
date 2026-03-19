package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// DestinationUnaryInterceptor создает клиентский унарный интерцептор,
// который добавляет заголовок 'x-destination-service' в каждый исходящий запрос.
func DestinationUnaryInterceptor(destinationService string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Создаем метаданные с нашим заголовком
		md := metadata.Pairs("x-destination-service", destinationService)

		// Прикрепляем их к существующему контексту
		newCtx := metadata.NewOutgoingContext(ctx, md)

		// Вызываем следующего в цепочке (или сам RPC-вызов), передавая новый контекст
		return invoker(newCtx, method, req, reply, cc, opts...)
	}
}

// DestinationStreamInterceptor - то же самое, но для стриминговых RPC.
// Даже если у вас их нет, хорошей практикой является реализация обоих интерцепторов.
func DestinationStreamInterceptor(destinationService string) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		md := metadata.Pairs("x-destination-service", destinationService)
		newCtx := metadata.NewOutgoingContext(ctx, md)
		return streamer(newCtx, desc, cc, method, opts...)
	}
}
