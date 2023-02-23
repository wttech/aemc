import org.apache.jackrabbit.oak.spi.security.user.util.PasswordUtil
import org.apache.jackrabbit.oak.spi.commit.CommitInfo
import org.apache.jackrabbit.oak.spi.commit.EmptyHook

class Global {
    static userNode = null;
}

void findUserNode(ub) {
    if (ub.hasProperty("rep:principalName")) {
        if ("rep:principalName = [[.User]]".equals(ub.getProperty("rep:principalName").toString())) {
            Global.userNode = ub;
        }
    }
    ub.childNodeNames.each { it ->
        if (Global.userNode == null) {
            findUserNode(ub.getChildNode(it));
        }
    }
}

ub = session.store.root.builder();
findUserNode(ub.getChildNode("home").getChildNode("users"));

if (Global.userNode) {
    println("Found user node: " + Global.userNode.toString());
    Global.userNode.setProperty("rep:password", PasswordUtil.buildPasswordHash("[[.Password]]"));
    session.store.merge(ub, EmptyHook.INSTANCE, CommitInfo.EMPTY);
    println("Updated user node: " + Global.userNode.toString());
} else {
    println("Could not find user node!");
}
