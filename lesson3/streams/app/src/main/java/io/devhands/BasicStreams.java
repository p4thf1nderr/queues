package io.devhands;

import org.apache.kafka.common.serialization.Serdes;
import org.apache.kafka.streams.KafkaStreams;
import org.apache.kafka.streams.StreamsBuilder;
import org.apache.kafka.streams.kstream.Consumed;
import org.apache.kafka.streams.kstream.GlobalKTable;
import org.apache.kafka.streams.kstream.KStream;
import org.apache.kafka.streams.kstream.Materialized;
import org.apache.kafka.streams.kstream.Produced;

import java.io.FileInputStream;
import java.io.IOException;
import java.time.Duration;
import java.util.Properties;
import java.util.concurrent.CountDownLatch;

public class BasicStreams {

    public static void main(String[] args) throws IOException {
        Properties props = new Properties();
        try (FileInputStream fis = new FileInputStream("streams.properties")) {
            props.load(fis);
        }

        props.put("application.id", "basic-streams");
        props.put("client.id", "basic-streams-client");
        props.put("consumer.group.instance.id", "consumer-id-1");

        final String sourceTopic = "product-events";
        final String joinTopic = "gift-table-events";
        final String outputTopic = "gift-events";

        StreamsBuilder builder = new StreamsBuilder();

        System.out.println("Consuming from topic [" + sourceTopic + "] and producing to [" + outputTopic + "] via "
                + props.get("bootstrap.servers"));

        // KTable хранит информацию какой подарок выдавать при заказе
        // определенной продукции
        GlobalKTable<String, String> giftsTable = builder.globalTable(
                joinTopic,
                Materialized.with(Serdes.String(), Serdes.String()));

        // из SourceTopiс поступают события о заказе продукции
        KStream<String, String> sourceStream = builder.stream(sourceTopic,
                Consumed.with(Serdes.String(), Serdes.String()));

        // джойним таблицу и стрим, чтобы определить нужно ли отправить событие
        // на списание подарка со склада и какой подарок списать со склада
        // чтобы добавить его к заказу

        sourceStream
                .peek((key, value) -> System.out.println("In >> key: " + key + ":\t" + value))
                .join(giftsTable, (key, value) -> key, (streamVal, tableVal) -> tableVal)
                .peek((key, value) -> System.out.println("Out << key: " + key + ":\t" + value))
                .to(outputTopic, Produced.with(Serdes.String(), Serdes.String()));

        try (KafkaStreams kafkaStreams = new KafkaStreams(builder.build(), props)) {
            final CountDownLatch shutdownLatch = new CountDownLatch(1);
            Runtime.getRuntime().addShutdownHook(new Thread(() -> {
                kafkaStreams.close(Duration.ofSeconds(1));
                shutdownLatch.countDown();
            }));
            try {
                kafkaStreams.start();
                shutdownLatch.await();
            } catch (Throwable e) {
                System.exit(1);
            }
        }
        System.exit(0);
    }
}
