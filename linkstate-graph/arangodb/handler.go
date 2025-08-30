package arangodb

import (
	"context"
	"fmt"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"

	"github.com/cisco-open/jalapeno/gobmp-arango/kafkanotifier"
	notifier "github.com/cisco-open/jalapeno/gobmp-arango/kafkanotifier"
	"github.com/sbezverk/gobmp/pkg/message"
)

func (a *arangoDB) lsNodeHandler(obj *notifier.EventMessage) error {
	ctx := context.TODO()
	if err := a.validateEventMessage(obj, a.lsnode.Name()); err != nil {
		return err
	}

	var node message.LSNode
	_, err := a.lsnode.ReadDocument(ctx, obj.Key, &node)

	switch {
	case driver.IsNotFoundGeneral(err) && obj.Action == "del":
		return a.processIgpNodeRemoval(ctx, obj.Key)
	case driver.IsNotFoundGeneral(err):
		return fmt.Errorf("document %s not found but Action is %s, possible stale event",
			obj.Key, obj.Action)
	case err != nil:
		return fmt.Errorf("failed to read document %s: %w", obj.Key, err)
	}

	switch obj.Action {
	case "add", "update":
		if err := a.processIgpNode(ctx, obj.Key, &node); err != nil {
			return fmt.Errorf("failed to process %s for vertex %s: %w",
				obj.Action, obj.Key, err)
		}
	}

	glog.V(5).Infof("Successfully processed %s action for node %s", obj.Action, obj.Key)
	return nil
}

func (a *arangoDB) lsSRv6SIDHandler(obj *notifier.EventMessage) error {
	ctx := context.TODO()
	if err := a.validateEventMessage(obj, a.lssrv6sid.Name()); err != nil {
		return err
	}

	var sid message.LSSRv6SID
	_, err := a.lssrv6sid.ReadDocument(ctx, obj.Key, &sid)

	switch {
	case driver.IsNotFoundGeneral(err) && obj.Action == "del":
		glog.V(6).Infof("SRv6 SID deleted: %s for igp_node key: %s", obj.Action, obj.Key)
		return nil
	case driver.IsNotFoundGeneral(err):
		return fmt.Errorf("document %s not found but Action is %s, possible stale event",
			obj.Key, obj.Action)
	case err != nil:
		return fmt.Errorf("failed to read document %s: %w", obj.Key, err)
	}

	switch obj.Action {
	case "add":
		if err := a.processLSSRv6SID(ctx, obj.Key, obj.ID, &sid); err != nil {
			return fmt.Errorf("failed to process %s for edge %s: %w",
				obj.Action, obj.Key, err)
		}
	}

	glog.V(6).Infof("Successfully processed %s action for SRv6 SID %s", obj.Action, obj.Key)
	return nil
}

func (a *arangoDB) lsPrefixHandler(obj *kafkanotifier.EventMessage) error {
	ctx := context.TODO()
	if err := a.validateEventMessage(obj, a.lsprefix.Name()); err != nil {
		return err
	}

	var prefix message.LSPrefix
	_, err := a.lsprefix.ReadDocument(ctx, obj.Key, &prefix)

	switch {
	case driver.IsNotFoundGeneral(err) && obj.Action == "del":
		if strings.Contains(obj.Key, ":") {
			return a.processv6PrefixRemoval(ctx, obj.Key, obj.Action)
		}
		return a.processPrefixRemoval(ctx, obj.Key, obj.Action)
	case driver.IsNotFoundGeneral(err):
		return fmt.Errorf("document %s not found but Action is %s, possible stale event",
			obj.Key, obj.Action)
	case err != nil:
		return fmt.Errorf("failed to read document %s: %w", obj.Key, err)
	}

	switch obj.Action {
	case "add", "update":
		if err := a.processPrefixSID(ctx, obj.Key, obj.ID, prefix); err != nil {
			return fmt.Errorf("failed to process %s for vertex %s: %w",
				obj.Action, obj.Key, err)
		}
		if err := a.processLSPrefixEdge(ctx, obj.Key, &prefix); err != nil {
			return fmt.Errorf("failed to process %s for edge %s: %w",
				obj.Action, obj.Key, err)
		}
	}

	glog.V(5).Infof("Successfully processed %s action for prefix %s", obj.Action, obj.Key)
	return nil
}

func (a *arangoDB) lsLinkHandler(obj *kafkanotifier.EventMessage) error {
	ctx := context.TODO()
	if err := a.validateEventMessage(obj, a.lslink.Name()); err != nil {
		return err
	}

	var link message.LSLink
	_, err := a.lslink.ReadDocument(ctx, obj.Key, &link)

	switch {
	case driver.IsNotFoundGeneral(err) && obj.Action == "del":
		if strings.Contains(obj.Key, ":") {
			return a.processv6LinkRemoval(ctx, obj.Key, obj.Action)
		}
		return a.processLinkRemoval(ctx, obj.Key, obj.Action)
	case driver.IsNotFoundGeneral(err):
		return fmt.Errorf("document %s not found but Action is %s, possible stale event",
			obj.Key, obj.Action)
	case err != nil:
		return fmt.Errorf("failed to read document %s: %w", obj.Key, err)
	}

	switch obj.Action {
	case "add", "update":
		if err := a.processLSLinkEdge(ctx, obj.Key, &link); err != nil {
			return fmt.Errorf("failed to process %s for edge %s: %w",
				obj.Action, obj.Key, err)
		}
	}

	glog.V(5).Infof("Successfully processed %s action for link %s", obj.Action, obj.Key)
	return nil
}

// Common validation function
func (a *arangoDB) validateEventMessage(obj *notifier.EventMessage, expectedCollection string) error {
	if obj == nil {
		return fmt.Errorf("event message is nil")
	}

	collection := strings.Split(obj.ID, "/")[0]
	if collection != expectedCollection {
		return fmt.Errorf("collection name mismatch: expected %s, got %s",
			expectedCollection, collection)
	}

	return nil
}
