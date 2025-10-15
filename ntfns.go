package bisonbotkit

import (
	"context"
	"errors"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
)

func (b *Bot) gcNtfns(ctx context.Context) error {
	var ackRes types.AckResponse
	var ackReq types.AckRequest
	for {
		// Keep requesting a new stream if the connection breaks.
		streamReq := types.GCMStreamRequest{UnackedFrom: ackReq.SequenceId}
		stream, err := b.chatService.GCMStream(ctx, &streamReq)
		if errors.Is(err, context.Canceled) {
			// Program is done.
			return err
		}
		if err != nil {
			b.gcLog.Errorf("failed to get GC stream: %v", err)
			time.Sleep(time.Second) // Wait to try again.
			continue
		}

		b.gcLog.Info("Listening for GC msgs...")
		for {
			var pm types.GCReceivedMsg
			err := stream.Recv(&pm)
			if errors.Is(err, context.Canceled) {
				// Program is done.
				return err
			}
			if err != nil {
				b.gcLog.Errorf("failed to receive from stream: %v", err)
				break
			}
			ackReq.SequenceId = pm.SequenceId
			err = b.chatService.AckReceivedGCM(ctx, &ackReq, &ackRes)
			if err != nil {
				b.gcLog.Errorf("failed to acknowledge received gc: %v", err)
				break
			}
			b.gcChan <- &pm
		}
	}
}

func (b *Bot) inviteNtfns(ctx context.Context) error {
	var ackRes types.AckResponse
	var ackReq types.AckRequest
	for {
		// Keep requesting a new stream if the connection breaks. Also
		// request any messages received since the last one we acked.
		streamReq := types.ReceivedGCInvitesRequest{UnackedFrom: ackReq.SequenceId}
		stream, err := b.gcService.ReceivedGCInvites(ctx, &streamReq)
		if errors.Is(err, context.Canceled) {
			// Program is done.
			return err
		}
		if err != nil {
			b.gcLog.Warnf("Error while obtaining GC invite stream: %v", err)
			time.Sleep(time.Second) // Wait to try again.
			continue
		}

		b.gcLog.Info("Listening for GC invites...")
		for {
			var pm types.ReceivedGCInvite
			err := stream.Recv(&pm)
			if errors.Is(err, context.Canceled) {
				// Program is done.
				return err
			}
			if err != nil {
				b.gcLog.Warnf("Error while receiving invite stream: %v", err)
				break
			}
			ackReq.SequenceId = pm.SequenceId
			if err = b.gcService.AckReceivedGCInvites(ctx, &ackReq, &ackRes); err != nil {
				b.gcLog.Errorf("failed to acknowledge kx: %v", err)
				break
			}
			b.inviteChan <- &pm
		}
	}
}

func (b *Bot) kxNtfns(ctx context.Context) error {
	var ksr types.KXStreamRequest
	var ackReq types.AckRequest
	var ackRes types.AckResponse
	for {
		stream, err := b.chatService.KXStream(ctx, &ksr)
		if errors.Is(err, context.Canceled) {
			// Program is done.
			return err
		}
		if err != nil {
			b.kxLog.Warnf("Error while obtaining KX stream: %v", err)
			time.Sleep(time.Second) // Wait to try again.
			continue
		}
		b.kxLog.Info("Listening for kxs...")
		for {
			var pm types.KXCompleted
			err := stream.Recv(&pm)
			if errors.Is(err, context.Canceled) {
				// Program is done.
				return err
			}
			if err != nil {
				b.kxLog.Warnf("Error while receiving stream: %v", err)
				break
			}
			ksr.UnackedFrom = pm.SequenceId
			ackReq.SequenceId = pm.SequenceId
			if err = b.chatService.AckKXCompleted(ctx, &ackReq, &ackRes); err != nil {
				b.kxLog.Errorf("failed to acknowledge kx: %v", err)
				break
			}
			b.kxChan <- &pm
		}
	}
}

func (b *Bot) pmNtfns(ctx context.Context) error {
	var ackRes types.AckResponse
	var ackReq types.AckRequest
	for {
		// Keep requesting a new stream if the connection breaks.
		streamReq := types.PMStreamRequest{UnackedFrom: ackReq.SequenceId}
		stream, err := b.chatService.PMStream(ctx, &streamReq)
		if errors.Is(err, context.Canceled) {
			// Program is done.
			return err
		}
		if err != nil {
			b.pmLog.Errorf("failed to get PM stream: %v", err)
			time.Sleep(time.Second) // Wait to try again.
			continue
		}

		b.pmLog.Info("Listening for private messages...")
		for {
			var pm types.ReceivedPM
			err := stream.Recv(&pm)
			if errors.Is(err, context.Canceled) {
				// Program is done.
				return err
			}
			if err != nil {
				b.pmLog.Errorf("failed to receive from stream: %v", err)
				break
			}
			ackReq.SequenceId = pm.SequenceId
			err = b.chatService.AckReceivedPM(ctx, &ackReq, &ackRes)
			if err != nil {
				b.pmLog.Errorf("failed to acknowledge received gc: %v", err)
				break
			}
			b.pmChan <- &pm
		}
	}
}

func (b *Bot) postNtfns(ctx context.Context) error {
	var psr types.PostsStreamRequest
	var ackReq types.AckRequest
	var ackRes types.AckResponse
	for {
		stream, err := b.postService.PostsStream(ctx, &psr)
		if errors.Is(err, context.Canceled) {
			// Program is done.
			return err
		}
		if err != nil {
			b.postLog.Errorf("failed to setup posts stream: %v", err)
			time.Sleep(time.Second) // Wait to try again.
			continue
		}
		b.postLog.Info("Listening for posts...")
		for {
			var pm types.ReceivedPost
			err := stream.Recv(&pm)
			if errors.Is(err, context.Canceled) {
				// Program is done.
				return err
			}
			if err != nil {
				b.postLog.Errorf("failed to receive from stream: %v", err)
				break
			}
			psr.UnackedFrom = pm.SequenceId
			ackReq.SequenceId = pm.SequenceId
			if err = b.postService.AckReceivedPost(ctx, &ackReq, &ackRes); err != nil {
				b.postLog.Errorf("failed to acknowledge post: %v", err)
				break
			}
			b.postChan <- &pm
		}
	}
}

func (b *Bot) postStatusNtfns(ctx context.Context) error {
	var psr types.PostsStatusStreamRequest
	var ackReq types.AckRequest
	var ackRes types.AckResponse
	for {
		stream, err := b.postService.PostsStatusStream(ctx, &psr)
		if errors.Is(err, context.Canceled) {
			// Program is done.
			return err
		}
		if err != nil {
			b.postStatusLog.Warnf("Error while obtaining posts status stream: %v", err)
			time.Sleep(time.Second) // Wait to try again.
			continue
		}
		b.postStatusLog.Info("Listening for comments...")
		for {
			var pm types.ReceivedPostStatus
			err := stream.Recv(&pm)
			if errors.Is(err, context.Canceled) {
				// Program is done.
				return err
			}
			if err != nil {
				b.postStatusLog.Warnf("Error while receiving posts status stream: %v", err)
				break
			}
			psr.UnackedFrom = pm.SequenceId
			ackReq.SequenceId = pm.SequenceId
			if err = b.postService.AckReceivedPostStatus(ctx, &ackReq, &ackRes); err != nil {
				b.postStatusLog.Errorf("Failed to acknowledge post status: %v", err)
				break
			}
			b.postStatusChan <- &pm
		}
	}
}

func (b *Bot) tipProgress(ctx context.Context) error {
	var tpr types.TipProgressRequest
	for {
		stream, err := b.paymentService.TipProgress(ctx, &tpr)
		if errors.Is(err, context.Canceled) {
			// Program is done.
			return err
		}
		if err != nil {
			b.tipProgressLog.Warnf("Error while creating tip progress stream: %v", err)
			time.Sleep(time.Second) // Wait to try again.
			continue
		}
		b.tipProgressLog.Info("Listening for tip progress...")
		for {
			var pm types.TipProgressEvent
			err := stream.Recv(&pm)
			if errors.Is(err, context.Canceled) {
				// Program is done.
				return err
			}
			if err != nil {
				b.tipProgressLog.Warnf("Error while receiving stream: %v", err)
				break
			}
			b.tipProgressChan <- &pm
		}
	}
}

func (b *Bot) tipReceived(ctx context.Context) error {
	var tpr types.TipStreamRequest
	for {
		stream, err := b.paymentService.TipStream(ctx, &tpr)
		if errors.Is(err, context.Canceled) {
			// Program is done.
			return err
		}
		if err != nil {
			b.tipReceivedLog.Warnf("Error while creating tip received stream: %v", err)
			time.Sleep(time.Second) // Wait to try again.
			continue
		}
		b.tipReceivedLog.Info("Listening for tip received...")
		for {
			var tip types.ReceivedTip
			err := stream.Recv(&tip)
			if errors.Is(err, context.Canceled) {
				// Program is done.
				return err
			}
			if err != nil {
				b.tipReceivedLog.Warnf("Error while receiving stream: %v", err)
				break
			}
			b.tipReceivedChan <- &tip
		}
	}
}
